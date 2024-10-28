package database

import (
	"context"
	"fmt"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupTTLIndex() error {
	collection := GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("errpr get index list: %v", err)
	}
	defer cursor.Close(ctx)

	indexExists := false
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return fmt.Errorf("index decode error: %v", err)
		}
		if index["name"] == "expires_at_1" {
			indexExists = true
			break
		}
	}

	if !indexExists {
		indexModel := mongo.IndexModel{
			Keys:    bson.M{"expires_at": 1},
			Options: options.Index().SetExpireAfterSeconds(0),
		}
		if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
			return fmt.Errorf("error create ttl index: %v", err)
		}
		fmt.Println("Successful create ttl index.")
	} else {
		fmt.Println("ttl index already exist.")
	}

	return nil
}
