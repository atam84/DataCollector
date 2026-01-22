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

// ConnectorRepository handles database operations for connectors
type ConnectorRepository struct {
	collection *mongo.Collection
}

// NewConnectorRepository creates a new connector repository
func NewConnectorRepository(db *Database) *ConnectorRepository {
	collection := db.GetCollection("connectors")

	// Create unique index on exchange_id
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "exchange_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = collection.Indexes().CreateOne(ctx, indexModel)

	return &ConnectorRepository{
		collection: collection,
	}
}

// Create inserts a new connector
func (r *ConnectorRepository) Create(ctx context.Context, connector *models.Connector) error {
	connector.ID = primitive.NewObjectID()
	connector.CreatedAt = time.Now()
	connector.UpdatedAt = time.Now()

	// Initialize rate limit if not set
	if connector.RateLimit.PeriodStart.IsZero() {
		connector.RateLimit.PeriodStart = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, connector)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("connector with exchange_id '%s' already exists", connector.ExchangeID)
		}
		return fmt.Errorf("failed to create connector: %w", err)
	}

	return nil
}

// FindByID retrieves a connector by its ID
func (r *ConnectorRepository) FindByID(ctx context.Context, id string) (*models.Connector, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connector ID: %w", err)
	}

	var connector models.Connector
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&connector)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("connector not found")
		}
		return nil, fmt.Errorf("failed to find connector: %w", err)
	}

	return &connector, nil
}

// FindByExchangeID retrieves a connector by exchange ID
func (r *ConnectorRepository) FindByExchangeID(ctx context.Context, exchangeID string) (*models.Connector, error) {
	var connector models.Connector
	err := r.collection.FindOne(ctx, bson.M{"exchange_id": exchangeID}).Decode(&connector)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("connector not found for exchange: %s", exchangeID)
		}
		return nil, fmt.Errorf("failed to find connector: %w", err)
	}

	return &connector, nil
}

// FindAll retrieves all connectors with optional filters
func (r *ConnectorRepository) FindAll(ctx context.Context, filter bson.M) ([]*models.Connector, error) {
	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find connectors: %w", err)
	}
	defer cursor.Close(ctx)

	var connectors []*models.Connector
	if err := cursor.All(ctx, &connectors); err != nil {
		return nil, fmt.Errorf("failed to decode connectors: %w", err)
	}

	return connectors, nil
}

// Update updates a connector
func (r *ConnectorRepository) Update(ctx context.Context, id string, update bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid connector ID: %w", err)
	}

	// Add updated_at timestamp
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update connector: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("connector not found")
	}

	return nil
}

// Delete removes a connector
func (r *ConnectorRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid connector ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete connector: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("connector not found")
	}

	return nil
}

// AcquireRateLimitToken attempts to acquire a rate limit token atomically
// DEPRECATED: Use RateLimiter.WaitForSlot() instead for proper throttling
func (r *ConnectorRepository) AcquireRateLimitToken(ctx context.Context, exchangeID string, weight int) (bool, error) {
	now := time.Now()

	// Find and update atomically
	filter := bson.M{
		"exchange_id": exchangeID,
		"status":      "active",
	}

	var connector models.Connector
	err := r.collection.FindOne(ctx, filter).Decode(&connector)
	if err != nil {
		return false, fmt.Errorf("failed to find connector: %w", err)
	}

	// Check if we need to reset the period
	periodElapsed := now.Sub(connector.RateLimit.PeriodStart).Milliseconds()
	if periodElapsed >= int64(connector.RateLimit.PeriodMs) {
		// Reset period
		update := bson.M{
			"rate_limit.usage":        weight,
			"rate_limit.period_start": now,
			"updated_at":              now,
		}

		_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
		return err == nil, err
	}

	// Check if we can acquire the token
	if connector.RateLimit.Usage+weight > connector.RateLimit.Limit {
		return false, nil // Rate limit exceeded
	}

	// Increment usage atomically
	update := bson.M{
		"$inc": bson.M{"rate_limit.usage": weight},
		"$set": bson.M{"updated_at": now},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("failed to acquire token: %w", err)
	}

	return result.ModifiedCount > 0, nil
}

// ResetRateLimitPeriod resets the rate limit period for an exchange
func (r *ConnectorRepository) ResetRateLimitPeriod(ctx context.Context, exchangeID string) error {
	now := time.Now()

	filter := bson.M{"exchange_id": exchangeID}
	update := bson.M{
		"$set": bson.M{
			"rate_limit.usage":        0,
			"rate_limit.period_start": now,
			"updated_at":              now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to reset rate limit period: %w", err)
	}

	return nil
}

// IncrementAPIUsage increments the API usage counter and updates last call timestamp
func (r *ConnectorRepository) IncrementAPIUsage(ctx context.Context, exchangeID string) error {
	now := time.Now()

	filter := bson.M{"exchange_id": exchangeID}
	update := bson.M{
		"$inc": bson.M{"rate_limit.usage": 1},
		"$set": bson.M{
			"rate_limit.last_api_call_at": now,
			"updated_at":                  now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increment API usage: %w", err)
	}

	return nil
}

// GetRateLimitStats returns rate limit statistics for a connector
func (r *ConnectorRepository) GetRateLimitStats(ctx context.Context, exchangeID string) (*models.RateLimit, error) {
	var connector models.Connector
	err := r.collection.FindOne(ctx, bson.M{"exchange_id": exchangeID}).Decode(&connector)
	if err != nil {
		return nil, fmt.Errorf("failed to find connector: %w", err)
	}

	return &connector.RateLimit, nil
}

// RecordSuccessfulCall records a successful API call and updates health metrics
func (r *ConnectorRepository) RecordSuccessfulCall(ctx context.Context, exchangeID string, responseTimeMs int64) error {
	now := time.Now()

	// First get the current connector to calculate new average
	var connector models.Connector
	err := r.collection.FindOne(ctx, bson.M{"exchange_id": exchangeID}).Decode(&connector)
	if err != nil {
		return fmt.Errorf("failed to find connector: %w", err)
	}

	// Calculate new average response time
	newTotalCalls := connector.Health.TotalCalls + 1
	newAvgResponseMs := ((connector.Health.AverageResponseMs * float64(connector.Health.TotalCalls)) + float64(responseTimeMs)) / float64(newTotalCalls)

	// Calculate uptime percentage
	uptimePercentage := float64(newTotalCalls-connector.Health.TotalFailures) / float64(newTotalCalls) * 100

	// Determine health status
	healthStatus := "healthy"
	if connector.Health.ConsecutiveFailures > 0 {
		// Just recovered from failure
		healthStatus = "healthy"
	}

	filter := bson.M{"exchange_id": exchangeID}
	update := bson.M{
		"$set": bson.M{
			"health.status":               healthStatus,
			"health.last_successful_call": now,
			"health.consecutive_failures": 0,
			"health.average_response_ms":  newAvgResponseMs,
			"health.last_response_ms":     responseTimeMs,
			"health.last_health_check":    now,
			"health.uptime_percentage":    uptimePercentage,
			"updated_at":                  now,
		},
		"$inc": bson.M{
			"health.total_calls": 1,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to record successful call: %w", err)
	}

	return nil
}

// RecordFailedCall records a failed API call and updates health metrics
func (r *ConnectorRepository) RecordFailedCall(ctx context.Context, exchangeID string, errorMsg string) error {
	now := time.Now()

	// First get the current connector
	var connector models.Connector
	err := r.collection.FindOne(ctx, bson.M{"exchange_id": exchangeID}).Decode(&connector)
	if err != nil {
		return fmt.Errorf("failed to find connector: %w", err)
	}

	newConsecutiveFailures := connector.Health.ConsecutiveFailures + 1
	newTotalCalls := connector.Health.TotalCalls + 1
	newTotalFailures := connector.Health.TotalFailures + 1

	// Calculate uptime percentage
	uptimePercentage := float64(newTotalCalls-newTotalFailures) / float64(newTotalCalls) * 100

	// Determine health status based on consecutive failures
	healthStatus := "healthy"
	if newConsecutiveFailures >= 5 {
		healthStatus = "unhealthy"
	} else if newConsecutiveFailures >= 2 {
		healthStatus = "degraded"
	}

	filter := bson.M{"exchange_id": exchangeID}
	update := bson.M{
		"$set": bson.M{
			"health.status":               healthStatus,
			"health.last_failed_call":     now,
			"health.last_error":           errorMsg,
			"health.consecutive_failures": newConsecutiveFailures,
			"health.last_health_check":    now,
			"health.uptime_percentage":    uptimePercentage,
			"updated_at":                  now,
		},
		"$inc": bson.M{
			"health.total_calls":    1,
			"health.total_failures": 1,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to record failed call: %w", err)
	}

	return nil
}

// GetHealthStatus returns the health status for a connector
func (r *ConnectorRepository) GetHealthStatus(ctx context.Context, exchangeID string) (*models.ConnectorHealth, error) {
	var connector models.Connector
	err := r.collection.FindOne(ctx, bson.M{"exchange_id": exchangeID}).Decode(&connector)
	if err != nil {
		return nil, fmt.Errorf("failed to find connector: %w", err)
	}

	return &connector.Health, nil
}

// GetAllHealthStatuses returns health status for all connectors
func (r *ConnectorRepository) GetAllHealthStatuses(ctx context.Context) (map[string]*models.ConnectorHealth, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find connectors: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]*models.ConnectorHealth)
	for cursor.Next(ctx) {
		var connector models.Connector
		if err := cursor.Decode(&connector); err != nil {
			continue
		}
		health := connector.Health
		result[connector.ExchangeID] = &health
	}

	return result, nil
}
