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

// AlertRepository handles database operations for alerts
type AlertRepository struct {
	collection       *mongo.Collection
	configCollection *mongo.Collection
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *Database) *AlertRepository {
	collection := db.GetCollection("alerts")
	configCollection := db.GetCollection("alert_config")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "severity", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "source.type", Value: 1},
				{Key: "source.id", Value: 1},
			},
		},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &AlertRepository{
		collection:       collection,
		configCollection: configCollection,
	}
}

// Create inserts a new alert
func (r *AlertRepository) Create(ctx context.Context, alert *models.Alert) error {
	alert.ID = primitive.NewObjectID()
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	if alert.Status == "" {
		alert.Status = models.AlertStatusActive
	}

	_, err := r.collection.InsertOne(ctx, alert)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// FindByID retrieves an alert by its ID
func (r *AlertRepository) FindByID(ctx context.Context, id string) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %w", err)
	}

	var alert models.Alert
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&alert)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("alert not found")
		}
		return nil, fmt.Errorf("failed to find alert: %w", err)
	}

	return &alert, nil
}

// FindAll retrieves all alerts with optional filters
func (r *AlertRepository) FindAll(ctx context.Context, filter bson.M, limit int64) ([]*models.Alert, error) {
	if filter == nil {
		filter = bson.M{}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find alerts: %w", err)
	}
	defer cursor.Close(ctx)

	var alerts []*models.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, fmt.Errorf("failed to decode alerts: %w", err)
	}

	return alerts, nil
}

// FindActive retrieves all active alerts
func (r *AlertRepository) FindActive(ctx context.Context) ([]*models.Alert, error) {
	filter := bson.M{"status": models.AlertStatusActive}
	return r.FindAll(ctx, filter, 0)
}

// FindByStatus retrieves alerts by status
func (r *AlertRepository) FindByStatus(ctx context.Context, status models.AlertStatus) ([]*models.Alert, error) {
	filter := bson.M{"status": status}
	return r.FindAll(ctx, filter, 0)
}

// FindBySource retrieves alerts by source
func (r *AlertRepository) FindBySource(ctx context.Context, sourceType, sourceID string) ([]*models.Alert, error) {
	filter := bson.M{
		"source.type": sourceType,
		"source.id":   sourceID,
	}
	return r.FindAll(ctx, filter, 0)
}

// FindByJobID retrieves alerts for a specific job
func (r *AlertRepository) FindByJobID(ctx context.Context, jobID string) ([]*models.Alert, error) {
	return r.FindBySource(ctx, "job", jobID)
}

// FindByExchangeID retrieves alerts for a specific exchange/connector
func (r *AlertRepository) FindByExchangeID(ctx context.Context, exchangeID string) ([]*models.Alert, error) {
	filter := bson.M{"source.exchange_id": exchangeID}
	return r.FindAll(ctx, filter, 0)
}

// Update updates an alert
func (r *AlertRepository) Update(ctx context.Context, id string, update bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid alert ID: %w", err)
	}

	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// Acknowledge acknowledges an alert
func (r *AlertRepository) Acknowledge(ctx context.Context, id string, acknowledgedBy string) error {
	now := time.Now()
	update := bson.M{
		"status":          models.AlertStatusAcknowledged,
		"acknowledged_at": now,
		"acknowledged_by": acknowledgedBy,
	}
	return r.Update(ctx, id, update)
}

// Resolve marks an alert as resolved
func (r *AlertRepository) Resolve(ctx context.Context, id string) error {
	now := time.Now()
	update := bson.M{
		"status":      models.AlertStatusResolved,
		"resolved_at": now,
	}
	return r.Update(ctx, id, update)
}

// ResolveBySource resolves all active alerts for a specific source
func (r *AlertRepository) ResolveBySource(ctx context.Context, sourceType, sourceID string) (int64, error) {
	now := time.Now()
	filter := bson.M{
		"source.type": sourceType,
		"source.id":   sourceID,
		"status":      models.AlertStatusActive,
	}
	update := bson.M{
		"$set": bson.M{
			"status":      models.AlertStatusResolved,
			"resolved_at": now,
			"updated_at":  now,
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve alerts: %w", err)
	}

	return result.ModifiedCount, nil
}

// Delete removes an alert
func (r *AlertRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid alert ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// DeleteOlderThan deletes alerts older than the specified duration
func (r *AlertRepository) DeleteOlderThan(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	filter := bson.M{
		"created_at": bson.M{"$lt": cutoff},
		"status":     models.AlertStatusResolved,
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old alerts: %w", err)
	}

	return result.DeletedCount, nil
}

// GetSummary returns a summary of alerts
func (r *AlertRepository) GetSummary(ctx context.Context) (*models.AlertSummary, error) {
	// Count by status
	activeCount, _ := r.collection.CountDocuments(ctx, bson.M{"status": models.AlertStatusActive})
	acknowledgedCount, _ := r.collection.CountDocuments(ctx, bson.M{"status": models.AlertStatusAcknowledged})
	totalCount, _ := r.collection.CountDocuments(ctx, bson.M{
		"status": bson.M{"$in": []models.AlertStatus{models.AlertStatusActive, models.AlertStatusAcknowledged}},
	})

	// Count by severity (only active/acknowledged)
	bySeverity := make(map[models.AlertSeverity]int64)
	severities := []models.AlertSeverity{
		models.AlertSeverityInfo,
		models.AlertSeverityWarning,
		models.AlertSeverityError,
		models.AlertSeverityCritical,
	}
	for _, severity := range severities {
		count, _ := r.collection.CountDocuments(ctx, bson.M{
			"severity": severity,
			"status":   bson.M{"$in": []models.AlertStatus{models.AlertStatusActive, models.AlertStatusAcknowledged}},
		})
		if count > 0 {
			bySeverity[severity] = count
		}
	}

	// Count by type (only active/acknowledged)
	byType := make(map[models.AlertType]int64)
	types := []models.AlertType{
		models.AlertTypeJobFailed,
		models.AlertTypeJobConsecFailures,
		models.AlertTypeConnectorDown,
		models.AlertTypeRateLimitExceeded,
		models.AlertTypeNoDataCollected,
		models.AlertTypeSystemError,
	}
	for _, alertType := range types {
		count, _ := r.collection.CountDocuments(ctx, bson.M{
			"type":   alertType,
			"status": bson.M{"$in": []models.AlertStatus{models.AlertStatusActive, models.AlertStatusAcknowledged}},
		})
		if count > 0 {
			byType[alertType] = count
		}
	}

	return &models.AlertSummary{
		Active:       activeCount,
		Acknowledged: acknowledgedCount,
		Total:        totalCount,
		BySeverity:   bySeverity,
		ByType:       byType,
	}, nil
}

// ExistsActiveAlert checks if an active alert already exists for a source
func (r *AlertRepository) ExistsActiveAlert(ctx context.Context, alertType models.AlertType, sourceType, sourceID string) (bool, error) {
	filter := bson.M{
		"type":        alertType,
		"source.type": sourceType,
		"source.id":   sourceID,
		"status":      models.AlertStatusActive,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check existing alert: %w", err)
	}

	return count > 0, nil
}

// GetConfig retrieves the alert configuration
func (r *AlertRepository) GetConfig(ctx context.Context) (*models.AlertConfig, error) {
	var config models.AlertConfig
	err := r.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default config if none exists
			return models.DefaultAlertConfig(), nil
		}
		return nil, fmt.Errorf("failed to get alert config: %w", err)
	}
	return &config, nil
}

// SaveConfig saves the alert configuration
func (r *AlertRepository) SaveConfig(ctx context.Context, config *models.AlertConfig) error {
	config.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	if config.ID != primitive.NilObjectID {
		filter["_id"] = config.ID
	}

	update := bson.M{"$set": config}

	_, err := r.configCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save alert config: %w", err)
	}

	return nil
}
