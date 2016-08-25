package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/strava/go.strava"
	"github.com/team-first/grand-tour/core"
	"github.com/team-first/grand-tour/web"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

const (
	defaultHost string = "localhost"
	defaultPort int    = 8080
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
	User     *core.User
	LoginUrl string
}

type appHandler func(http.ResponseWriter, *http.Request) error

// https://blog.golang.org/error-handling-and-go
func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

type logHandler struct {
	handler http.Handler
}

func LogHandler(h http.Handler) http.Handler {
	return logHandler{h}
}

func (h logHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s %s", r.Method, r.URL))
	h.handler.ServeHTTP(w, r)
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

	if config.Secret == "" {
		err = errors.New("secret is required!")
	} else if config.Strava.Id == 0 {
		err = errors.New("[strava] id is required!")
	} else if config.Strava.Secret == "" {
		err = errors.New("[strava] secret is required!")
	}

	return config, err
}

func getPage(r *http.Request) (Page, error) {
	var page Page
	var err error

	page.LoginUrl = authenticator.AuthorizationURL("state1", strava.Permissions.Public, true)

	page.User, err = web.GetCurrentUser(r)

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

	web.InitStore([]byte(config.Secret))

	strava.ClientId = config.Strava.Id
	strava.ClientSecret = config.Strava.Secret

	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://%s:%d/callback", config.Host, config.Port),
		RequestClientGenerator: nil,
	}

	path, err := authenticator.CallbackPath()

	router := mux.NewRouter()
	router.Handle(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))
	router.Handle("/", appHandler(indexHandler))
	router.Handle("/logout", appHandler(logoutHandler))
	router.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	server := &http.Server{
		Handler:      LogHandler(router),
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println(fmt.Sprintf("Listening at http://%s:%d", config.Host, config.Port))
	log.Fatal(server.ListenAndServe())
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
	err := func() (err error) {
		err = createUser(auth.Athlete.Id)

		if err != nil {
			return err
		}

		user := &core.User{auth.Athlete.Id, auth.Athlete.FirstName}
		web.Login(w, r, user)

		page, err := getPage(r)

		if err != nil {
			return err
		}

		err = render(w, "success.html", page)

		return err
	}()

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	log.Println(err)

	page, err := getPage(r)

	if err != nil {
		log.Fatal(err)
	}

	// TODO show error message
	err = render(w, "failure.html", page)

	if err != nil {
		log.Fatal(err)
	}
}

// TODO only POST
func logoutHandler(w http.ResponseWriter, r *http.Request) (err error) {
	err = web.Logout(w, r)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/", http.StatusFound)

	return err
}
