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
