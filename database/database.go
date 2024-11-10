package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/godror/godror"
	"github.com/redis/go-redis/v9"
)

var DB *sql.DB
var RedisClient *redis.Client

func InitDatabase() (*sql.DB, *redis.Client, error) {
	var ctx = context.Background()

	// Initialize Redis client and assign it to the global RedisClient variable
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "mypassword",
		DB:       0,
	})

	// Test Redis connection
	status, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalln("Redis connection was refused:", err)
	}
	fmt.Println("Redis status:", status)

	// Initialize Oracle DB connection
	dsn := `user="system" password="Ahay1234" connectString="localhost:1521/FREE" timezone="Europe/Berlin"`
	DB, err = sql.Open("godror", dsn)
	if err != nil {
		return nil, nil, err
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(time.Hour)

	// Ensure connection is alive
	if err = DB.Ping(); err != nil {
		return nil, nil, err
	}

	// Create tables if they do not exist
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS TODOLIST (
			id         INTEGER GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) NOT NULL PRIMARY KEY,
			title      VARCHAR2(255),
			description VARCHAR2(255),
			status     VARCHAR2(50),
			due_date   DATE
		)
	`)
	if err != nil && !isTableAlreadyExistsError(err) {
		return nil, nil, err
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS USERS (
			id       INTEGER GENERATED ALWAYS AS IDENTITY (START WITH 1 INCREMENT BY 1) NOT NULL PRIMARY KEY,
			username VARCHAR2(50) UNIQUE NOT NULL,
			password VARCHAR2(70) NOT NULL
		)
	`)
	if err != nil && !isTableAlreadyExistsError(err) {
		return nil, nil, err
	}

	return DB, RedisClient, nil
}

// Helper function to identify if the table already exists
func isTableAlreadyExistsError(err error) bool {
	return err != nil && err.Error() == "ORA-00955" // ORA-00955 is Oracle's code for "name is already used by an existing object"
}
