package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"html/template"
	"log"
	"net/http"
	"os"
)

const defaultPort int = 8080

type Config struct {
	Id     int
	Secret string
	Port   int
}

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

	if config.Port == 0 {
		config.Port = defaultPort
	}

	return config, err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, nil)
}

func main() {
	var configFilename = flag.String("conf", "config.toml", "path to config file")
	flag.Parse()

	config, err := readConfig(*configFilename)

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
