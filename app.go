package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/strava/go.strava"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	defaultHost string = "localhost"
	defaultPort int    = 8080
	sessionName string = "grand-tour"
	templateDir string = "templates"
)

type Config struct {
	Host   string
	Port   int
	Secret string
	Strava Strava
}

type Strava struct {
	Id     int
	Secret string
}

type Page struct {
	User     *User
	LoginUrl string
}

type User struct {
	Id        int64
	FirstName string
}

type handler func(http.ResponseWriter, *http.Request) error

// https://blog.golang.org/error-handling-and-go
func (f handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

var authenticator *strava.OAuthAuthenticator
var db *sql.DB
var store *sessions.CookieStore

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

	if config.Secret == "" {
		err = errors.New("secret is required!")
	} else if config.Strava.Id == 0 {
		err = errors.New("[strava] id is required!")
	} else if config.Strava.Secret == "" {
		err = errors.New("[strava] secret is required!")
	}

	return config, err
}

func getCurrentUser(r *http.Request) (user *User, err error) {
	session, err := store.Get(r, sessionName)

	if err != nil {
		return user, err
	}

	id, ok := session.Values["id"].(int64)

	if ok {
		user = new(User)
		user.Id = id
	}

	return user, err
}

func getPage(r *http.Request) (Page, error) {
	var page Page
	var err error

	page.LoginUrl = authenticator.AuthorizationURL("state1", strava.Permissions.Public, true)

	page.User, err = getCurrentUser(r)

	if err != nil {
		return page, err
	}

	return page, err
}

func render(w http.ResponseWriter, filename string, params interface{}) (err error) {
	bp := path.Join(templateDir, "base.html")
	fp := path.Join(templateDir, filename)

	t, err := template.ParseFiles(bp, fp)

	if err != nil {
		return err
	}

	t.ExecuteTemplate(w, "base", params)

	return err
}

func indexHandler(w http.ResponseWriter, r *http.Request) (err error) {
	log.Println("In indexhandler")

	page, err := getPage(r)

	if err != nil {
		return err
	}

	err = render(w, "index.html", page)

	return err
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

	store = sessions.NewCookieStore([]byte(config.Secret))

	strava.ClientId = config.Strava.Id
	strava.ClientSecret = config.Strava.Secret

	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://%s:%d/callback", config.Host, config.Port),
		RequestClientGenerator: nil,
	}

	path, err := authenticator.CallbackPath()

	http.HandleFunc(path,authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))
	http.Handle("/", handler(indexHandler))
	http.Handle("/logout", handler(logoutHandler))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	log.Println(fmt.Sprintf("Listening at http://%s:%d", config.Host, config.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), context.ClearHandler(http.DefaultServeMux)))
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
	var err error

	err = createUser(auth.Athlete.Id)

	if err != nil {
		// TODO
		log.Fatal(err)
	}

	session, err := store.Get(r, sessionName)

	if err != nil {
		// TODO
		log.Fatal(err)
	}

	session.Values["id"] = auth.Athlete.Id
	session.Save(r, w)

	page, err := getPage(r)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("OAuthSuccess")

	err = render(w, "success.html", page)

	if err != nil {
		// TODO
		log.Fatal(err)
	}
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	page, err := getPage(r)

	if err != nil {
		log.Fatal(err)
	}

	// TODO show error message
	log.Println("OAuth Failure")
	err = render(w, "failure.html", page)

	if err != nil {
		log.Fatal(err)
	}
}

// TODO only POST
func logoutHandler(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := store.Get(r, sessionName)

	if err != nil {
		return err
	}

	delete(session.Values, "id")

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)

	return err
}
