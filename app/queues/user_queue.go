package queues

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

func GetUserQueueByEmail(email string) (models.User, error) {
	var user models.User
	err := database.GetCollection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	return user, err
}

func GetUserQueueByUsername(username string) (models.User, error) {
	var user models.User
	err := database.GetCollection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	return user, err
}

func GetUserQueueByID(id uint64) (models.User, error) {
	var user models.User
	err := database.GetCollection("users").FindOne(context.Background(), bson.M{"id": id}).Decode(&user)
	return user, err
}

func CreateUserQueue(user models.User) error {
	_, err := database.GetCollection("users").InsertOne(context.Background(), user)
	return err
}

func AppendUserProviderQueue(id uint64, user models.User) error {
	filter := bson.M{"id": id}

	update := bson.M{
		"$set": bson.M{
			"providers":  user.Providers,
			"updated_at": user.UpdatedAt,
		},
	}

	_, err := database.GetCollection("users").UpdateOne(context.Background(), filter, update)
	return err
}
