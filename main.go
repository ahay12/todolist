package main

import (
	"log"
	"todolist/database"
	"todolist/router"
)

func main() {
	_, rdb, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}

	defer func() {
		if err := rdb.Close(); err != nil {
			log.Fatal("Failed to close database: ", err)
		}
	}()

	// Initialize router and start the server
	app, logFile := router.Make() // Make function returns the app and the log file
	defer logFile.Close()         // Close the log file when the application exits

	// Start the server in a separate goroutine
	go func() {
		if err := app.Listen(":4000"); err != nil {
			log.Fatalf("Failed to start the server: %v", err)
		}
	}()

	// Gracefully handle shutdown
	select {} // Block forever, server continues running
}
