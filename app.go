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

var authenticator *strava.OAuthAuthenticator

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

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/success.html")
	t.Execute(w, nil)
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	// TODO show error message
	t, _ := template.ParseFiles("templates/failure.html")
	t.Execute(w, nil)
}
