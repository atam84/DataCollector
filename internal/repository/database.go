package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database wraps the MongoDB client and provides access to collections
type Database struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect establishes a connection to MongoDB
func Connect(uri, dbName string) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := &Database{
		Client:   client,
		Database: client.Database(dbName),
	}

	return db, nil
}

// Close disconnects from MongoDB
func (db *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Client.Disconnect(ctx)
}

// HealthCheck verifies the database connection is alive
func (db *Database) HealthCheck(ctx context.Context) error {
	return db.Client.Ping(ctx, readpref.Primary())
}

// GetCollection returns a MongoDB collection
func (db *Database) GetCollection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}
