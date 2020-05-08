package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"time"
)

func main () {
	log.Println("Starting consumer")

	/*** INITIALIZE CONSUMER ***/
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id": "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatal(err)
	}

	//err = consumer.SubscribeTopics([]string{"cars"}, nil)
	if err = consumer.Subscribe("announces", nil); err != nil {
		log.Fatal(err)
	}

	/*** INITIALIZE MONGODB CLIENT ***/
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Panic(err)
	}

	client.StartSession()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err = client.Connect(ctx); err !=nil {
		log.Panic(err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Panic(err)
	}

	collection := client.Database("reezorcar").Collection("announces")

	/*** HANDLE MESSAGES ***/
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			//fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))

			// Handle message on a go routine
			go handleStandalone(string(msg.Value), collection)
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

