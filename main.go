package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"leadgentracker/internals/handler"
	"leadgentracker/internals/repository"
	"leadgentracker/internals/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	dbClient := configureDatabaseConnection()
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			log.Fatal("could not disconnect from database: ", err)
		}
	}()

	sseBroadcaster := handler.NewSSEBroadcaster()
	leadHandler := configureLeadHandler(dbClient, sseBroadcaster)
	configureEndpointHandlers(leadHandler, sseBroadcaster)

	// start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func configureEndpointHandlers(leadHandler *handler.LeadHandler, sseBroadcaster *handler.SSEBroadcaster) {
	http.HandleFunc("/", leadHandler.ServeIndex)
	http.HandleFunc("/add-lead", leadHandler.AddLead)
	http.HandleFunc("/update-lead", leadHandler.UpdateLead)
	http.HandleFunc("/delete-lead", leadHandler.DeleteLead)
	http.HandleFunc("/lead-stats", leadHandler.GetLeadStats)
	http.HandleFunc("/leads", leadHandler.GetAllLeads)
	http.HandleFunc("/sse", sseBroadcaster.HandleSSE)
}

func configureLeadHandler(client *mongo.Client, sseBroadcaster *handler.SSEBroadcaster) *handler.LeadHandler {
	leadRepo := repository.NewLeadRepository(client)
	statsRepo := repository.NewStatsRepository(client)

	statsService := service.NewStatsService(statsRepo)
	leadService := service.NewLeadService(leadRepo)

	return handler.NewLeadHandler(leadService, statsService, sseBroadcaster)
}

func configureDatabaseConnection() *mongo.Client {
	username := os.Getenv("MONGO_USER")
	password := os.Getenv("MONGO_PASSWORD")
	host := os.Getenv("MONGO_HOST")
	port := os.Getenv("MONGO_PORT")
	database := os.Getenv("MONGO_DB")

	log.Printf("Database configuration: host=%s, port=%s, database=%s, username=%s",
		host, port, database, username)

	if username == "" || password == "" || host == "" || port == "" || database == "" {
		log.Fatal("environment variables missing")
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", username, password, host, port, database)
	log.Printf("connecting to database %s...", database)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("could not connect to MongoDB: ", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("could not ping MongoDB: ", err)
	}

	log.Println("successfully connected")
	return client
}
