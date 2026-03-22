package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Initialize all schemas + indexes
func InitSchemas() {
	createUserProfile()
	createUserAuth()
	createWallet()
	createTransaction()
	createSession()
}

// ---------------- USER PROFILE ----------------

func createUserProfile() {
	collectionName := "userprofile"

	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"user_id", "first_name", "created_at"},
			"properties": bson.M{
				"user_id":    bson.M{"bsonType": "string"},
				"first_name": bson.M{"bsonType": "string"},
				"last_name":  bson.M{"bsonType": "string"},
				"created_at": bson.M{"bsonType": "date"},
				"updated_at": bson.M{"bsonType": "date"},
			},
		},
	}

	createCollection(collectionName, validator)

	col := GetCollection(collectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: optionsUnique(),
	}

	createIndex(col, indexModel)
}

// ---------------- USER AUTH ----------------

func createUserAuth() {
	collectionName := "userauth"

	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"user_id", "email", "hash", "created_at"},
			"properties": bson.M{
				"user_id": bson.M{"bsonType": "string"},
				"email": bson.M{
					"bsonType": "string",
					"pattern":  "^.+@.+\\..+$",
				},
				"hash":       bson.M{"bsonType": "string"},
				"created_at": bson.M{"bsonType": "date"},
				"updated_at": bson.M{"bsonType": "date"},
			},
		},
	}

	createCollection(collectionName, validator)

	col := GetCollection(collectionName)

	createIndex(col, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: optionsUnique(),
	})

	createIndex(col, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: optionsUnique(),
	})
}

// ---------------- WALLET ----------------

func createWallet() {
	collectionName := "wallet"

	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"user_id", "balance", "currency", "created_at"},
			"properties": bson.M{
				"user_id": bson.M{"bsonType": "string"},
				"balance": bson.M{
					"bsonType": "string",
				},
				"currency":   bson.M{"bsonType": "string"},
				"created_at": bson.M{"bsonType": "date"},
				"updated_at": bson.M{"bsonType": "date"},
			},
		},
	}

	createCollection(collectionName, validator)

	col := GetCollection(collectionName)

	createIndex(col, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: optionsUnique(),
	})
}

// ---------------- TRANSACTION ----------------

func createTransaction() {
	collectionName := "transaction"

	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"transaction_id", "user_id", "amount", "type"},
			"properties": bson.M{
				"transaction_id": bson.M{"bsonType": "string"},
				"user_id":        bson.M{"bsonType": "string"},
				"currency":       bson.M{"bsonType": "string"},
				"amount":         bson.M{"bsonType": "double"},
				"type": bson.M{
					"enum": []string{"credit", "debit", "rollback"},
				},
				"request_id":    bson.M{"bsonType": "string"},
				"balance_after": bson.M{"bsonType": "string"},
				"created_at":    bson.M{"bsonType": "date"},
				"updated_at":    bson.M{"bsonType": "date"},
			},
		},
	}

	createCollection(collectionName, validator)

	col := GetCollection(collectionName)

	createIndex(col, mongo.IndexModel{
		Keys:    bson.D{{Key: "transaction_id", Value: 1}},
		Options: optionsUnique(),
	})

	createIndex(col, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
}

// ---------------- SESSION ----------------

func createSession() {
	collectionName := "session"

	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{
				"session_id", "user_id", "operator_id",
				"game_id", "currency", "timestamp", "created_at",
			},
			"properties": bson.M{
				"session_id":  bson.M{"bsonType": "string"},
				"user_id":     bson.M{"bsonType": "string"},
				"operator_id": bson.M{"bsonType": "string"},
				"game_id":     bson.M{"bsonType": "string"},
				"balance":     bson.M{"bsonType": "string"},
				"currency":    bson.M{"bsonType": "string"},
				"timestamp":   bson.M{"bsonType": "long"},
				"created_at":  bson.M{"bsonType": "date"},
				"updated_at":  bson.M{"bsonType": "date"},
			},
		},
	}

	createCollection(collectionName, validator)

	col := GetCollection(collectionName)

	createIndex(col, mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: optionsUnique(),
	})

	createIndex(col, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
}

// ---------------- COMMON HELPERS ----------------

func createCollection(name string, validator bson.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Client.Database("casino-lobby").CreateCollection(ctx, name)
	if err != nil {
		// collection may already exist → update validator
		cmd := bson.D{
			{"collMod", name},
			{"validator", validator},
			{"validationLevel", "moderate"},
		}
		err = Client.Database("casino-lobby").RunCommand(ctx, cmd).Err()
		if err != nil {
			log.Println("Validator update failed for", name, err)
		}
		return
	}

	// apply validator
	cmd := bson.D{
		{"collMod", name},
		{"validator", validator},
	}
	Client.Database("casino-lobby").RunCommand(ctx, cmd)
}

func createIndex(col *mongo.Collection, model mongo.IndexModel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := col.Indexes().CreateOne(ctx, model)
	if err != nil {
		log.Println("Index creation failed:", err)
	}
}

func optionsUnique() *options.IndexOptions {
	return options.Index().SetUnique(true)
}
