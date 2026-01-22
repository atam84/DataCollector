package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yourusername/datacollector/internal/models"
)

// JobRepository handles database operations for jobs
type JobRepository struct {
	collection *mongo.Collection
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *Database) *JobRepository {
	collection := db.GetCollection("jobs")

	// Create unique compound index on (exchange_id, symbol, timeframe)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "connector_exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	// Create index on status and next_run_time for scheduler queries
	statusIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "run_state.next_run_time", Value: 1},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = collection.Indexes().CreateOne(ctx, indexModel)
	_, _ = collection.Indexes().CreateOne(ctx, statusIndexModel)

	return &JobRepository{
		collection: collection,
	}
}

// Create inserts a new job
func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	job.ID = primitive.NewObjectID()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	// Set default status if not provided
	if job.Status == "" {
		job.Status = "active"
	}

	// Initialize schedule mode if not set
	if job.Schedule.Mode == "" {
		job.Schedule.Mode = "timeframe"
	}

	_, err := r.collection.InsertOne(ctx, job)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("job already exists for %s/%s/%s", job.ConnectorExchangeID, job.Symbol, job.Timeframe)
		}
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// FindByID retrieves a job by its ID
func (r *JobRepository) FindByID(ctx context.Context, id string) (*models.Job, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	var job models.Job
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	return &job, nil
}

// FindAll retrieves all jobs with optional filters
func (r *JobRepository) FindAll(ctx context.Context, filter bson.M) ([]*models.Job, error) {
	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	return jobs, nil
}

// FindByConnector retrieves all jobs for a specific connector
func (r *JobRepository) FindByConnector(ctx context.Context, exchangeID string) ([]*models.Job, error) {
	filter := bson.M{"connector_exchange_id": exchangeID}
	return r.FindAll(ctx, filter)
}

// FindRunnableJobs retrieves jobs that are ready to run
func (r *JobRepository) FindRunnableJobs(ctx context.Context) ([]*models.Job, error) {
	now := time.Now()

	filter := bson.M{
		"status": "active",
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"run_state.next_run_time": bson.M{"$lte": now}},
					{"run_state.next_run_time": nil},
				},
			},
			{
				"$or": []bson.M{
					{"run_state.locked_until": bson.M{"$lte": now}},
					{"run_state.locked_until": nil},
				},
			},
		},
	}

	return r.FindAll(ctx, filter)
}

// Update updates a job
func (r *JobRepository) Update(ctx context.Context, id string, update bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	// Add updated_at timestamp
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// UpdateStatus updates the job status
func (r *JobRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	update := bson.M{"status": status}
	return r.Update(ctx, id, update)
}

// UpdateCursor updates the job cursor
func (r *JobRepository) UpdateCursor(ctx context.Context, id string, lastCandleTime time.Time) error {
	update := bson.M{
		"cursor.last_candle_time": lastCandleTime,
	}
	return r.Update(ctx, id, update)
}

// AcquireLock attempts to acquire a lock on a job for execution
func (r *JobRepository) AcquireLock(ctx context.Context, id string, duration time.Duration) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("invalid job ID: %w", err)
	}

	now := time.Now()
	lockUntil := now.Add(duration)

	// Only lock if not currently locked or lock has expired
	filter := bson.M{
		"_id": objectID,
		"$or": []bson.M{
			{"run_state.locked_until": bson.M{"$lte": now}},
			{"run_state.locked_until": nil},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"run_state.locked_until": lockUntil,
			"updated_at":             now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return result.ModifiedCount > 0, nil
}

// ReleaseLock releases the lock on a job
func (r *JobRepository) ReleaseLock(ctx context.Context, id string) error {
	update := bson.M{
		"run_state.locked_until": nil,
	}
	return r.Update(ctx, id, update)
}

// RecordRun updates job state after a run
func (r *JobRepository) RecordRun(ctx context.Context, id string, success bool, nextRunTime *time.Time, errorMsg *string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	now := time.Now()

	setFields := bson.M{
		"run_state.last_run_time": now,
		"run_state.locked_until":  nil,
		"updated_at":              now,
	}

	if nextRunTime != nil {
		setFields["run_state.next_run_time"] = *nextRunTime
	}

	if errorMsg != nil {
		setFields["run_state.last_error"] = *errorMsg
		if !success {
			setFields["status"] = "error"
		}
	} else {
		setFields["run_state.last_error"] = nil
	}

	update := bson.M{
		"$set": setFields,
		"$inc": bson.M{"run_state.runs_total": 1},
	}

	fmt.Printf("[DEBUG] RecordRun for job %s: update=%+v\n", id, update)

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)

	if err != nil {
		fmt.Printf("[DEBUG] RecordRun error: %v\n", err)
		return fmt.Errorf("failed to record run: %w", err)
	}

	fmt.Printf("[DEBUG] RecordRun result: matched=%d, modified=%d\n", result.MatchedCount, result.ModifiedCount)

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// Delete removes a job
func (r *JobRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// CountByConnector counts jobs for a specific connector
func (r *JobRepository) CountByConnector(ctx context.Context, exchangeID string) (int64, error) {
	filter := bson.M{"connector_exchange_id": exchangeID}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count jobs: %w", err)
	}
	return count, nil
}

// CountActiveByConnector counts active jobs for a specific connector
func (r *JobRepository) CountActiveByConnector(ctx context.Context, exchangeID string) (int64, error) {
	filter := bson.M{
		"connector_exchange_id": exchangeID,
		"status":                "active",
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count active jobs: %w", err)
	}
	return count, nil
}

// UpdateStatusByConnector updates status for all jobs of a connector
func (r *JobRepository) UpdateStatusByConnector(ctx context.Context, exchangeID string, status string) error {
	filter := bson.M{"connector_exchange_id": exchangeID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update jobs status: %w", err)
	}

	return nil
}

// FixMissingNextRunTime sets next_run_time for jobs that don't have it
func (r *JobRepository) FixMissingNextRunTime(ctx context.Context) (int64, error) {
	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"run_state.next_run_time": nil},
			{"run_state.next_run_time": bson.M{"$exists": false}},
		},
	}

	nextRunTime := time.Now().Add(1 * time.Minute)

	update := bson.M{
		"$set": bson.M{
			"run_state.next_run_time": nextRunTime,
			"updated_at":              time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to fix jobs: %w", err)
	}

	return result.ModifiedCount, nil
}

// IncrementConsecutiveFailures increments the consecutive failure counter
func (r *JobRepository) IncrementConsecutiveFailures(ctx context.Context, jobID string) error {
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	now := time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$inc": bson.M{
			"run_state.consecutive_failures": 1,
		},
		"$set": bson.M{
			"run_state.last_failure_time": now,
			"updated_at":                  now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increment consecutive failures: %w", err)
	}

	return nil
}

// ResetConsecutiveFailures resets the consecutive failure counter to 0
func (r *JobRepository) ResetConsecutiveFailures(ctx context.Context, jobID string) error {
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"run_state.consecutive_failures": 0,
			"updated_at":                     time.Now(),
		},
		"$unset": bson.M{
			"run_state.last_failure_time": "",
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to reset consecutive failures: %w", err)
	}

	return nil
}

// GetJobsWithFailures returns jobs that have consecutive failures
func (r *JobRepository) GetJobsWithFailures(ctx context.Context) ([]*models.Job, error) {
	filter := bson.M{
		"status": "active",
		"run_state.consecutive_failures": bson.M{"$gt": 0},
	}

	opts := options.Find().SetSort(bson.D{{Key: "run_state.consecutive_failures", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs with failures: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	return jobs, nil
}

// SetDependencies sets the dependencies for a job
func (r *JobRepository) SetDependencies(ctx context.Context, jobID string, dependsOn []primitive.ObjectID) error {
	objectID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	update := bson.M{
		"depends_on": dependsOn,
		"updated_at": time.Now(),
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to set dependencies: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// FindByIDs retrieves multiple jobs by their IDs
func (r *JobRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Job, error) {
	if len(ids) == 0 {
		return []*models.Job{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	return jobs, nil
}

// GetDependencyStatus returns the dependency status for a job
func (r *JobRepository) GetDependencyStatus(ctx context.Context, jobID string, maxAge time.Duration) (*models.DependencyStatus, error) {
	job, err := r.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	status := &models.DependencyStatus{
		JobID:            jobID,
		DependsOn:        make([]string, 0),
		BlockedBy:        make([]string, 0),
		AllDepsCompleted: true,
	}

	if len(job.DependsOn) == 0 {
		return status, nil
	}

	// Convert to string IDs
	for _, depID := range job.DependsOn {
		status.DependsOn = append(status.DependsOn, depID.Hex())
	}

	// Get all dependency jobs
	depJobs, err := r.FindByIDs(ctx, job.DependsOn)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency jobs: %w", err)
	}

	now := time.Now()
	cutoffTime := now.Add(-maxAge)

	// Check which dependencies have completed recently
	for _, depJob := range depJobs {
		completedRecently := depJob.RunState.LastRunTime != nil &&
			depJob.RunState.LastRunTime.After(cutoffTime) &&
			depJob.RunState.LastError == nil

		if !completedRecently {
			status.BlockedBy = append(status.BlockedBy, depJob.ID.Hex())
			status.AllDepsCompleted = false
		}
	}

	// Check for missing dependencies (deleted jobs)
	if len(depJobs) != len(job.DependsOn) {
		status.AllDepsCompleted = false
	}

	return status, nil
}

// CheckCircularDependency checks if adding a dependency would create a cycle
func (r *JobRepository) CheckCircularDependency(ctx context.Context, jobID string, newDepID string) (bool, error) {
	targetID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return false, fmt.Errorf("invalid job ID: %w", err)
	}

	newDepObjID, err := primitive.ObjectIDFromHex(newDepID)
	if err != nil {
		return false, fmt.Errorf("invalid dependency ID: %w", err)
	}

	// A job cannot depend on itself
	if jobID == newDepID {
		return true, nil
	}

	// BFS to check for cycles
	visited := make(map[primitive.ObjectID]bool)
	queue := []primitive.ObjectID{newDepObjID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if currentID == targetID {
			return true, nil // Found a cycle
		}

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Get the current job and its dependencies
		currentJob, err := r.FindByID(ctx, currentID.Hex())
		if err != nil {
			continue // Job might not exist
		}

		queue = append(queue, currentJob.DependsOn...)
	}

	return false, nil
}

// FindJobsDependingOn returns jobs that depend on the given job ID
func (r *JobRepository) FindJobsDependingOn(ctx context.Context, jobID string) ([]*models.Job, error) {
	objectID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	filter := bson.M{
		"depends_on": objectID,
	}

	return r.FindAll(ctx, filter)
}
