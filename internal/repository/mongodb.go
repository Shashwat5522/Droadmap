package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bacancy/droadmap/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository handles MongoDB operations for tenant databases
type MongoRepository struct {
	client *mongo.Client
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(connString string) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("unable to ping MongoDB: %w", err)
	}

	return &MongoRepository{client: client}, nil
}

// CreateTenantDatabase creates a new database for a tenant (by creating a collection)
func (r *MongoRepository) CreateTenantDatabase(ctx context.Context, tenantName string) error {
	// MongoDB creates databases automatically when you write to them
	// We'll create an empty collection to ensure the database exists
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	db := r.client.Database(dbName)
	
	// Create the documents collection
	err := db.CreateCollection(ctx, "documents")
	if err != nil {
		// Collection might already exist, which is okay
		// Check if it's a different error
		if !mongo.IsDuplicateKeyError(err) && err.Error() != "Collection already exists" {
			return fmt.Errorf("unable to create collection: %w", err)
		}
	}

	// Create indexes
	collection := db.Collection("documents")
	_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_name", Value: 1}, {Key: "uploaded_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "file_name", Value: 1}},
		},
	})

	return err
}

// InsertDocument inserts a document into the tenant's database
func (r *MongoRepository) InsertDocument(ctx context.Context, tenantName string, doc *models.Document) error {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	collection := r.client.Database(dbName).Collection("documents")

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("unable to insert document: %w", err)
	}

	doc.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// TenantDatabaseExists checks if a tenant database exists
func (r *MongoRepository) TenantDatabaseExists(ctx context.Context, tenantName string) (bool, error) {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	
	// List all databases
	databases, err := r.client.ListDatabaseNames(ctx, bson.M{"name": dbName})
	if err != nil {
		return false, err
	}

	return len(databases) > 0, nil
}

// CountDocuments counts active (non-deleted) documents in a tenant database
func (r *MongoRepository) CountDocuments(ctx context.Context, tenantName string) (int64, error) {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	collection := r.client.Database(dbName).Collection("documents")
	
	// Only count non-deleted documents
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("unable to count documents: %w", err)
	}
	
	return count, nil
}

// SoftDeleteAllDocuments marks all documents in a tenant database as deleted
func (r *MongoRepository) SoftDeleteAllDocuments(ctx context.Context, tenantName string) (int64, error) {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	collection := r.client.Database(dbName).Collection("documents")
	
	// Update all documents to mark them as deleted
	filter := bson.M{"is_deleted": bson.M{"$ne": true}} // Only update non-deleted docs
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("unable to soft delete documents: %w", err)
	}
	
	return result.ModifiedCount, nil
}

// RestoreAllDocuments restores all soft-deleted documents in a tenant database
func (r *MongoRepository) RestoreAllDocuments(ctx context.Context, tenantName string) (int64, error) {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	collection := r.client.Database(dbName).Collection("documents")
	
	// Update all documents to restore them
	filter := bson.M{"is_deleted": true}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": false,
		},
		"$unset": bson.M{
			"deleted_at": "",
		},
	}
	
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("unable to restore documents: %w", err)
	}
	
	return result.ModifiedCount, nil
}

// DropDatabase drops a tenant database completely (for hard delete)
func (r *MongoRepository) DropDatabase(ctx context.Context, tenantName string) error {
	dbName := fmt.Sprintf("tenant_%s", tenantName)
	
	err := r.client.Database(dbName).Drop(ctx)
	if err != nil {
		return fmt.Errorf("unable to drop database: %w", err)
	}
	
	return nil
}

// Close closes the MongoDB connection
func (r *MongoRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

