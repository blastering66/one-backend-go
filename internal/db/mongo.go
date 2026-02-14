// Package db provides MongoDB connection and index management.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connect establishes a connection to MongoDB and returns the database handle.
func Connect(ctx context.Context, uri, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	slog.Info("connected to MongoDB", "db", dbName)
	return client.Database(dbName), nil
}

// EnsureIndexes creates required indexes idempotently.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// ── Users ──────────────────────────────────────────────────────────
	usersCol := db.Collection("users")
	_, err := usersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("db: index users.email: %w", err)
	}

	// ── Products ───────────────────────────────────────────────────────
	productsCol := db.Collection("products")
	productIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "description", Value: "text"},
			},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err = productsCol.Indexes().CreateMany(ctx, productIndexes)
	if err != nil {
		return fmt.Errorf("db: index products: %w", err)
	}

	// ── Refresh Tokens ─────────────────────────────────────────────────
	rtCol := db.Collection("refresh_tokens")
	rtIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "token", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	}
	_, err = rtCol.Indexes().CreateMany(ctx, rtIndexes)
	if err != nil {
		return fmt.Errorf("db: index refresh_tokens: %w", err)
	}

	slog.Info("database indexes ensured")
	return nil
}

// Disconnect gracefully closes the MongoDB connection.
func Disconnect(ctx context.Context, db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.Client().Disconnect(ctx)
}
