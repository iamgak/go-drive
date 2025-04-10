package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iamgak/go-drive/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Model           *models.Init
	UserID          uint
	isAuthenticated bool
	Email           string
	Logger          *logrus.Logger
	BaseDir         string
}

func main() {
	var logrusLogger = logrus.New()
	logrusLogger.SetFormatter(&logrus.TextFormatter{}) // Use JSON format for structured logging
	logrusLogger.SetLevel(logrus.InfoLevel)            // Log Info, Warning, and Error

	logrusLogger.Info("Task Web App startet \n")
	err := godotenv.Load()
	if err != nil {
		logrusLogger.Error("Failed to get task: ", err)
		log.Fatal("Error loading .env file:", err)
	}

	Port := os.Getenv("PORT")
	addr := flag.String("addr", ":"+Port, "HTTP network address")
	flag.Parse()

	dbORM, err := openDBORM()
	if err != nil {
		logrusLogger.Error("Error creating db connection : ", err)
		log.Fatal(err)
	}

	app := Application{
		Model:  models.Constructor(dbORM, logrusLogger),
		Logger: logrusLogger,
	}

	MigrateDB(dbORM)

	maxHeaderBytes := 1 << 20
	server := &http.Server{
		Addr:           *addr,
		Handler:        app.InitRouter(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	logrusLogger.Info("start http server listening ", *addr)
	server.ListenAndServe()
}
