package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/driver"
	"github.com/faebut/prozessor_gostack/internal/handlers"
	"github.com/faebut/prozessor_gostack/internal/helpers"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/render"
	"github.com/faebut/prozessor_gostack/internal/slack"
	"github.com/joho/godotenv"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

	defer close(app.MailChan)

	log.Println("Starting mail listener...")
	// start listening for mail
	listenForMail()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	portNumber := os.Getenv("PORT")

	log.Println("Starting application on port", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Visitor{})
	gob.Register(models.Role{})
	// gob.Register(map[string]int{})

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	// read env
	inProduction, err := strconv.ParseBool(os.Getenv("PRODUCTION"))
	if err != nil {
		log.Fatal(err)
	}
	useCache, err := strconv.ParseBool(os.Getenv("CACHE"))
	if err != nil {
		log.Fatal(err)
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

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 4 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to database...")
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
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
