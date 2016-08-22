package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/strava/go.strava"
	"html/template"
	"log"
	"net/http"
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultHost string = "localhost"
	defaultPort int    = 8080
)

type Config struct {
	Id     int
	Secret string
	Host   string
	Port   int
}

type Params struct {
	Url string
}

type User struct {
	FirstName string
}

var authenticator *strava.OAuthAuthenticator
var db *sql.DB

func readConfig(filename string) (Config, error) {
	var config Config
	var err error

	_, err = os.Stat(filename)

	if err != nil {
		return config, err
	}

	_, err = toml.DecodeFile(filename, &config)

	if err != nil {
		return config, err
	}

	if config.Host == "" {
		config.Host = defaultHost
	}

	if config.Port == 0 {
		config.Port = defaultPort
	}

	return config, err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	url := authenticator.AuthorizationURL("state1", strava.Permissions.Public, true)
	params := &Params{Url: url}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, params)
}

func main() {
	var configFilename = flag.String("conf", "config.toml", "path to config file")
	flag.Parse()

	config, err := readConfig(*configFilename)

	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("sqlite3", "./app.db")

	if err != nil {
		log.Fatal(err)
	}

	strava.ClientId = config.Id
	strava.ClientSecret = config.Secret

	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://%s:%d/callback", config.Host, config.Port),
		RequestClientGenerator: nil,
	}

	path, err := authenticator.CallbackPath()

	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}

func createUser(id int64) error {
	var exists bool
	var err error

	err = db.QueryRow("select exists(select 1 from users where id = ?)", id).Scan(&exists)

	if err != nil {
		return err
	}

	// Create the user if they don't exist
	// TODO race condition
	if !exists {
		statement, err := db.Prepare("insert into users (id) values (?)")

		if err != nil {
			return err
		}

		// Create the user
		_, err = statement.Exec(id)

		if err != nil {
			return err
		}
	}

	return err
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/success.html")

	err := createUser(auth.Athlete.Id)

	if err != nil {
		// TODO
		log.Fatal(err)
	}

	user := &User{FirstName: auth.Athlete.FirstName}

	t.Execute(w, user)
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	// TODO show error message
	t, _ := template.ParseFiles("templates/failure.html")
	t.Execute(w, nil)
}
