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

	update := bson.M{
		"run_state.last_run_time":  now,
		"run_state.locked_until":   nil,
		"updated_at":               now,
		"$inc": bson.M{"run_state.runs_total": 1},
	}

	if nextRunTime != nil {
		update["run_state.next_run_time"] = *nextRunTime
	}

	if errorMsg != nil {
		update["run_state.last_error"] = *errorMsg
		if !success {
			update["status"] = "error"
		}
	} else {
		update["run_state.last_error"] = nil
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to record run: %w", err)
	}

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
