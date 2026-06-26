package main

import (
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/driver"
	"github.com/faebut/prozessor_gostack/internal/events"
	"github.com/faebut/prozessor_gostack/internal/handlers"
	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/slack"
	"github.com/joho/godotenv"
)

var (
	app     config.AppConfig
	session *scs.SessionManager
)

// var infoLog *log.Logger
// var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		logger.Fatal("error starting app", slog.String("error", err.Error()))
	}

	defer db.SQL.Close()

	slog.Info("Warming up events cache...")
	go func() {
		if _, err := events.GetCalendarEvents(); err != nil {
			slog.Error("warming event cache", "error", err)
		}
	}()

	defer close(app.MailChan)

	slog.Info("Starting mail listener...")
	// start listening for mail
	listenForMail()

	err = godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	portNumber := os.Getenv("PORT")

	slog.Info("Starting application on port", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	logger.Fatal("listen and serve", slog.String("error", err.Error()))
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Visitor{})
	gob.Register(models.Role{})
	gob.Register(models.LogisticRequestSession{})
	// gob.Register(map[string]int{})

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	// read env
	inProduction, err := strconv.ParseBool(os.Getenv("PRODUCTION"))
	if err != nil {
		logger.Fatal("parsing .env file", slog.String("error", err.Error()))
	}
	useCache, err := strconv.ParseBool(os.Getenv("CACHE"))
	if err != nil {
		logger.Fatal("parsing .env file", slog.String("error", err.Error()))
	}
	dbHost := os.Getenv("DBHOST")
	dbName := os.Getenv("DBNAME")
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbPort := os.Getenv("DBPORT")
	dbSSL := os.Getenv("DBSSL")

	if dbName == "" || dbUser == "" {
		fmt.Println("Missing required entries in .env")
		os.Exit(1)
	}

	slackToken := os.Getenv("SLACKTOKEN")

	app.Slack = slack.NewSlack(slackToken)

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// change this to true when in production
	app.InProduction = inProduction
	app.UseCache = useCache
	//
	// infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// app.InfoLog = infoLog
	//
	// errorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	// app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 4 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	slog.Info("Connecting to database...")
	connectionString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPass,
		dbSSL,
	)

	// connect to db
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		logger.Fatal("connecting to database", slog.String("error", err.Error()))
	}
	slog.Info("Connected to database")

	// Initialize dual logging system
	coloredHandler := logger.NewColoredHandler(os.Stdout, slog.LevelInfo)
	dbHandler := logger.NewDBHandler(db.SQL, coloredHandler)
	slogLogger := slog.New(dbHandler)
	slog.SetDefault(slogLogger)

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)

	return db, nil
}
