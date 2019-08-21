package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/gorilla/mux"
)

var views = template.Must(template.ParseGlob("templates/*.html"))

type Service struct {
	Stage   string
	Version string
}

func main() {
	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()
	log.SetHandler(jsonhandler.Default)
	app.HandleFunc("/", index)
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	serviceInfo := []Service{
		Service{Stage: "foobar", Version: "123"},
	}
	err := views.ExecuteTemplate(w, "index.html", struct {
		Frontend []Service
	}{
		serviceInfo,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
