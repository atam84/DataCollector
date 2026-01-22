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

// IndicatorConfigRepository handles database operations for indicator configurations
type IndicatorConfigRepository struct {
	collection *mongo.Collection
}

// NewIndicatorConfigRepository creates a new indicator config repository
func NewIndicatorConfigRepository(db *Database) *IndicatorConfigRepository {
	collection := db.GetCollection("indicator_configs")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create unique index on name
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = collection.Indexes().CreateOne(ctx, indexModel)

	// Create index on is_default
	defaultIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "is_default", Value: 1}},
	}
	_, _ = collection.Indexes().CreateOne(ctx, defaultIndex)

	return &IndicatorConfigRepository{
		collection: collection,
	}
}

// Create inserts a new indicator configuration
func (r *IndicatorConfigRepository) Create(ctx context.Context, config *models.IndicatorConfig) error {
	config.ID = primitive.NewObjectID()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// If this is marked as default, unset other defaults
	if config.IsDefault {
		_, _ = r.collection.UpdateMany(ctx, bson.M{"is_default": true}, bson.M{"$set": bson.M{"is_default": false}})
	}

	_, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("config with name '%s' already exists", config.Name)
		}
		return fmt.Errorf("failed to create indicator config: %w", err)
	}

	return nil
}

// FindByID retrieves a configuration by its ID
func (r *IndicatorConfigRepository) FindByID(ctx context.Context, id string) (*models.IndicatorConfig, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	var config models.IndicatorConfig
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("config not found")
		}
		return nil, fmt.Errorf("failed to find config: %w", err)
	}

	return &config, nil
}

// FindByName retrieves a configuration by name
func (r *IndicatorConfigRepository) FindByName(ctx context.Context, name string) (*models.IndicatorConfig, error) {
	var config models.IndicatorConfig
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("config not found")
		}
		return nil, fmt.Errorf("failed to find config: %w", err)
	}

	return &config, nil
}

// FindDefault retrieves the default configuration
func (r *IndicatorConfigRepository) FindDefault(ctx context.Context) (*models.IndicatorConfig, error) {
	var config models.IndicatorConfig
	err := r.collection.FindOne(ctx, bson.M{"is_default": true}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return the built-in default config if none is set
			return models.DefaultIndicatorConfig(), nil
		}
		return nil, fmt.Errorf("failed to find default config: %w", err)
	}

	return &config, nil
}

// FindAll retrieves all configurations
func (r *IndicatorConfigRepository) FindAll(ctx context.Context) ([]*models.IndicatorConfig, error) {
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find configs: %w", err)
	}
	defer cursor.Close(ctx)

	var configs []*models.IndicatorConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("failed to decode configs: %w", err)
	}

	return configs, nil
}

// Update updates an indicator configuration
func (r *IndicatorConfigRepository) Update(ctx context.Context, id string, update bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	update["updated_at"] = time.Now()

	// If setting this as default, unset other defaults
	if isDefault, ok := update["is_default"].(bool); ok && isDefault {
		_, _ = r.collection.UpdateMany(ctx, bson.M{"is_default": true}, bson.M{"$set": bson.M{"is_default": false}})
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("config not found")
	}

	return nil
}

// Delete removes an indicator configuration
func (r *IndicatorConfigRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("config not found")
	}

	return nil
}

// SetDefault sets a configuration as the default
func (r *IndicatorConfigRepository) SetDefault(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	// Unset all other defaults
	_, err = r.collection.UpdateMany(ctx, bson.M{"is_default": true}, bson.M{"$set": bson.M{"is_default": false, "updated_at": time.Now()}})
	if err != nil {
		return fmt.Errorf("failed to unset defaults: %w", err)
	}

	// Set this as default
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"is_default": true, "updated_at": time.Now()}},
	)

	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("config not found")
	}

	return nil
}

// EnsureDefaultExists creates the default config if none exists
func (r *IndicatorConfigRepository) EnsureDefaultExists(ctx context.Context) error {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count configs: %w", err)
	}

	if count == 0 {
		defaultConfig := models.DefaultIndicatorConfig()
		return r.Create(ctx, defaultConfig)
	}

	return nil
}
