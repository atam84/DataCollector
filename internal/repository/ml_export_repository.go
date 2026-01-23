package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/datacollector/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MLExportRepository handles ML export persistence
type MLExportRepository struct {
	configCollection  *mongo.Collection
	jobCollection     *mongo.Collection
	datasetCollection *mongo.Collection
}

// NewMLExportRepository creates a new ML export repository
func NewMLExportRepository(db *Database) *MLExportRepository {
	configCollection := db.GetCollection("ml_export_configs")
	jobCollection := db.GetCollection("ml_export_jobs")
	datasetCollection := db.GetCollection("ml_datasets")

	// Create indexes for configs
	configCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_preset", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_default", Value: 1}},
		},
	})

	// Create indexes for jobs
	jobCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "job_ids", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
		},
	})

	// Create indexes for datasets
	datasetCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	})

	return &MLExportRepository{
		configCollection:  configCollection,
		jobCollection:     jobCollection,
		datasetCollection: datasetCollection,
	}
}

// ==================== Config Methods ====================

// CreateConfig creates a new export configuration
func (r *MLExportRepository) CreateConfig(ctx context.Context, config *models.MLExportConfig) error {
	if config.ID.IsZero() {
		config.ID = primitive.NewObjectID()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	_, err := r.configCollection.InsertOne(ctx, config)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("config with name '%s' already exists", config.Name)
	}
	return err
}

// FindConfigByID finds a config by ID
func (r *MLExportRepository) FindConfigByID(ctx context.Context, id primitive.ObjectID) (*models.MLExportConfig, error) {
	var config models.MLExportConfig
	err := r.configCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("config not found")
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindConfigs returns all user configs (not presets)
func (r *MLExportRepository) FindConfigs(ctx context.Context) ([]*models.MLExportConfig, error) {
	cursor, err := r.configCollection.Find(ctx, bson.M{"is_preset": bson.M{"$ne": true}}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []*models.MLExportConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// FindConfigByName finds a config by name
func (r *MLExportRepository) FindConfigByName(ctx context.Context, name string) (*models.MLExportConfig, error) {
	var config models.MLExportConfig
	err := r.configCollection.FindOne(ctx, bson.M{"name": name}).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindAllConfigs returns all configs
func (r *MLExportRepository) FindAllConfigs(ctx context.Context) ([]models.MLExportConfig, error) {
	cursor, err := r.configCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []models.MLExportConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// FindPresetConfigs returns only preset configs
func (r *MLExportRepository) FindPresetConfigs(ctx context.Context) ([]models.MLExportConfig, error) {
	cursor, err := r.configCollection.Find(ctx, bson.M{"is_preset": true}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []models.MLExportConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// FindUserConfigs returns user-created configs (not presets)
func (r *MLExportRepository) FindUserConfigs(ctx context.Context) ([]models.MLExportConfig, error) {
	cursor, err := r.configCollection.Find(ctx, bson.M{"is_preset": false}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []models.MLExportConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// FindDefaultConfig returns the default config
func (r *MLExportRepository) FindDefaultConfig(ctx context.Context) (*models.MLExportConfig, error) {
	var config models.MLExportConfig
	err := r.configCollection.FindOne(ctx, bson.M{"is_default": true}).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfig updates a config
func (r *MLExportRepository) UpdateConfig(ctx context.Context, config *models.MLExportConfig) error {
	config.UpdatedAt = time.Now()

	result, err := r.configCollection.UpdateOne(ctx, bson.M{"_id": config.ID}, bson.M{"$set": config})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("config not found")
	}
	return nil
}

// SetDefaultConfig sets a config as default (unsets others)
func (r *MLExportRepository) SetDefaultConfig(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	// Unset current default
	_, err = r.configCollection.UpdateMany(ctx, bson.M{"is_default": true}, bson.M{"$set": bson.M{"is_default": false}})
	if err != nil {
		return err
	}

	// Set new default
	result, err := r.configCollection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"is_default": true, "updated_at": time.Now()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("config not found")
	}
	return nil
}

// DeleteConfig deletes a config
func (r *MLExportRepository) DeleteConfig(ctx context.Context, id primitive.ObjectID) error {
	// Don't allow deleting presets
	var config models.MLExportConfig
	err := r.configCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("config not found")
	}
	if err != nil {
		return err
	}
	if config.IsPreset {
		return fmt.Errorf("cannot delete preset configs")
	}

	result, err := r.configCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("config not found")
	}
	return nil
}

// InitializePresets creates built-in presets if they don't exist
func (r *MLExportRepository) InitializePresets(ctx context.Context) error {
	presets := models.GetBuiltinPresets()

	for _, preset := range presets {
		existing, err := r.FindConfigByName(ctx, preset.Name)
		if err != nil {
			return err
		}
		if existing == nil {
			preset.ID = primitive.NewObjectID()
			preset.CreatedAt = time.Now()
			preset.UpdatedAt = time.Now()
			if _, err := r.configCollection.InsertOne(ctx, preset); err != nil {
				return err
			}
		}
	}
	return nil
}

// ==================== Export Job Methods ====================

// CreateExportJob creates a new export job
func (r *MLExportRepository) CreateExportJob(ctx context.Context, job *models.MLExportJob) error {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	_, err := r.jobCollection.InsertOne(ctx, job)
	return err
}

// FindExportJobByID finds an export job by ID
func (r *MLExportRepository) FindExportJobByID(ctx context.Context, id primitive.ObjectID) (*models.MLExportJob, error) {
	var job models.MLExportJob
	err := r.jobCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("export job not found")
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// FindExportJobs finds export jobs with pagination
func (r *MLExportRepository) FindExportJobs(ctx context.Context, limit, offset int64) ([]*models.MLExportJob, int64, error) {
	// Count total
	total, err := r.jobCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.jobCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var jobs []*models.MLExportJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// FindExportJobsByStatus finds jobs by status
func (r *MLExportRepository) FindExportJobsByStatus(ctx context.Context, status models.MLExportStatus) ([]models.MLExportJob, error) {
	cursor, err := r.jobCollection.Find(ctx, bson.M{"status": status}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []models.MLExportJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// FindRecentExportJobs finds recent export jobs
func (r *MLExportRepository) FindRecentExportJobs(ctx context.Context, limit int) ([]models.MLExportJob, error) {
	cursor, err := r.jobCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []models.MLExportJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// FindActiveExportJobs finds running or pending jobs
func (r *MLExportRepository) FindActiveExportJobs(ctx context.Context) ([]models.MLExportJob, error) {
	filter := bson.M{
		"status": bson.M{"$in": []string{string(models.MLExportStatusPending), string(models.MLExportStatusRunning)}},
	}
	cursor, err := r.jobCollection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []models.MLExportJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// UpdateExportJob updates an export job
func (r *MLExportRepository) UpdateExportJob(ctx context.Context, job *models.MLExportJob) error {
	job.UpdatedAt = time.Now()

	result, err := r.jobCollection.UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{"$set": job})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("export job not found")
	}
	return nil
}

// UpdateExportJobProgress updates job progress
func (r *MLExportRepository) UpdateExportJobProgress(ctx context.Context, id primitive.ObjectID, progress float64, processed int64, phase string) error {
	update := bson.M{
		"$set": bson.M{
			"progress":          progress,
			"processed_records": processed,
			"current_phase":     phase,
			"updated_at":        time.Now(),
		},
	}

	result, err := r.jobCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("export job not found")
	}
	return nil
}

// UpdateExportJobStatus updates job status
func (r *MLExportRepository) UpdateExportJobStatus(ctx context.Context, id primitive.ObjectID, status models.MLExportStatus, err string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if err != "" {
		update["$set"].(bson.M)["last_error"] = err
		update["$push"] = bson.M{"errors": err}
	}

	if status == models.MLExportStatusRunning {
		now := time.Now()
		update["$set"].(bson.M)["started_at"] = now
	}

	if status == models.MLExportStatusCompleted || status == models.MLExportStatusFailed {
		now := time.Now()
		update["$set"].(bson.M)["completed_at"] = now
	}

	result, errUpdate := r.jobCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if errUpdate != nil {
		return errUpdate
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("export job not found")
	}
	return nil
}

// CompleteExportJob marks job as complete with results
func (r *MLExportRepository) CompleteExportJob(ctx context.Context, id primitive.ObjectID, outputPath string, outputFiles []string, fileSize int64, featureCount int, columnNames []string, rowCount int64, metadata models.MLExportMetadata) error {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // Files expire after 24 hours

	update := bson.M{
		"$set": bson.M{
			"status":          models.MLExportStatusCompleted,
			"progress":        100.0,
			"output_path":     outputPath,
			"output_files":    outputFiles,
			"file_size_bytes": fileSize,
			"feature_count":   featureCount,
			"column_names":    columnNames,
			"row_count":       rowCount,
			"metadata":        metadata,
			"completed_at":    now,
			"expires_at":      expiresAt,
			"updated_at":      now,
		},
	}

	result, err := r.jobCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("export job not found")
	}
	return nil
}

// DeleteExportJob deletes an export job
func (r *MLExportRepository) DeleteExportJob(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.jobCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("export job not found")
	}
	return nil
}

// FindExpiredExportJobs finds jobs that have expired
func (r *MLExportRepository) FindExpiredExportJobs(ctx context.Context) ([]models.MLExportJob, error) {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"status":     models.MLExportStatusCompleted,
	}
	cursor, err := r.jobCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []models.MLExportJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// ==================== Dataset Methods ====================

// CreateDataset creates a new dataset
func (r *MLExportRepository) CreateDataset(ctx context.Context, dataset *models.MLDataset) error {
	if dataset.ID.IsZero() {
		dataset.ID = primitive.NewObjectID()
	}
	dataset.CreatedAt = time.Now()
	dataset.UpdatedAt = time.Now()

	_, err := r.datasetCollection.InsertOne(ctx, dataset)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("dataset with name '%s' already exists", dataset.Name)
	}
	return err
}

// FindDatasetByID finds a dataset by ID
func (r *MLExportRepository) FindDatasetByID(ctx context.Context, id string) (*models.MLDataset, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid dataset ID: %w", err)
	}

	var dataset models.MLDataset
	err = r.datasetCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&dataset)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// FindAllDatasets returns all datasets
func (r *MLExportRepository) FindAllDatasets(ctx context.Context) ([]models.MLDataset, error) {
	cursor, err := r.datasetCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var datasets []models.MLDataset
	if err := cursor.All(ctx, &datasets); err != nil {
		return nil, err
	}
	return datasets, nil
}

// UpdateDataset updates a dataset
func (r *MLExportRepository) UpdateDataset(ctx context.Context, dataset *models.MLDataset) error {
	dataset.UpdatedAt = time.Now()

	result, err := r.datasetCollection.UpdateOne(ctx, bson.M{"_id": dataset.ID}, bson.M{"$set": dataset})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("dataset not found")
	}
	return nil
}

// DeleteDataset deletes a dataset
func (r *MLExportRepository) DeleteDataset(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid dataset ID: %w", err)
	}

	result, err := r.datasetCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("dataset not found")
	}
	return nil
}
