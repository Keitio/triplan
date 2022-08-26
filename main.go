package main

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func getMongo() *mongo.Client {
	mongourl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		panic("env variable MONGO_URL must be present")
	}
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongourl))
	if err != nil {
		panic(err)
	}
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	return client
}

func main() {
	db := getMongo()
	defer func() {
		if err := db.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	collection := db.Database("stats").Collection("http_calls")
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		res := collection.FindOneAndUpdate(context.TODO(), bson.M{"_id": "/"}, bson.M{"$inc": bson.M{"count": 1}}, options.FindOneAndUpdate().SetUpsert(true))

		if res.Err() != nil {
			return c.JSON(fiber.Map{
				"error": "😢 could not do it: " + res.Err().Error(),
			})
		}
		out := map[string]any{}
		err := res.Decode(&out)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"message": "Wow, triplan !",
			"calls":   out["count"],
		})
	})

	app.Listen("0.0.0.0" + getPort())
}
