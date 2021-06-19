package main

import (
	"encoder/application/services"
	"encoder/framework/database"
	"encoder/framework/queue"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var db database.Database

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	autoMigrateDb, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		log.Fatalf("Error parsing boolean env var AUTO_MIGRATE_DB")
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Fatalf("Error parsing boolean env var DEBUG")
	}

	db.Dsn = os.Getenv("DSN")
	db.DbType = os.Getenv("DB_TYPE")
	db.DsnTest = os.Getenv("DSN_TEST")
	db.DbTypeTest = os.Getenv("DB_TYPE_TEST")
	db.Env = os.Getenv("ENV")
	db.AutoMigrateDb = autoMigrateDb
	db.Debug = debug
}

func main() {
	messageChannel := make(chan amqp.Delivery)
	jobReturnChannel := make(chan services.JobWorkerResult)

	db, err := db.Connect()
	if err != nil {
		log.Fatalf("Error connecting to DB")
	}

	defer db.Close()

	rabbitMQ := queue.NewRabbitMQ()
	channel := rabbitMQ.Connect()

	defer channel.Close()

	rabbitMQ.Consume(messageChannel)

	jobManager := services.NewJobManager(db, rabbitMQ, messageChannel, jobReturnChannel)
	jobManager.Start(channel)
}
