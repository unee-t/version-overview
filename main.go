package main

import (
	"bufio"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/gorilla/mux"
)

var views = template.Must(template.ParseGlob("templates/*.html"))
var commitRE = regexp.MustCompile(`<!-- COMMIT: (?P<version>\w+) -->`)

type Service struct {
	Site    string
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
	mefeVersions, err := getCOMMIT([]Service{
		{Site: "https://case.dev.unee-t.com"},
		{Site: "https://case.demo.unee-t.com"},
		{Site: "https://case.unee-t.com"},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = views.ExecuteTemplate(w, "index.html", struct {
		Frontend []Service
	}{
		mefeVersions,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getCOMMIT(check []Service) (version []Service, err error) {
	for i, v := range check {
		log.WithField("site", v).Info("getting version")
		resp, err := http.Get(v.Site)
		if err != nil {
			return check, err
		}
		defer resp.Body.Close()
		check[i].Version, err = versionString(resp.Body)
		if err != nil {
			return check, err
		}
	}
	return check, nil
}

func versionString(input io.Reader) (version string, err error) {
	// <!-- COMMIT: ae5b321 -->
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		txt := scanner.Text()
		if commitRE.MatchString(txt) {
			words := strings.Split(strings.TrimSpace(txt), " ")
			if len(words) > 3 {
				return words[2], nil
			}
		}
	}
	return "foo", nil
}
