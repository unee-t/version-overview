package main

import (
	"bufio"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/gorilla/mux"
)

var views = template.Must(template.ParseGlob("templates/*.html"))

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
		check[i].Version, err = commitVersion(resp.Body)
		if err != nil {
			return check, err
		}
	}
	return check, nil
}

func commitVersion(input io.Reader) (version string, err error) {
	// Version string is actually a commit id, e.g. "ae5b321" from:
	// <!-- COMMIT: ae5b321 -->
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		html := scanner.Text()
		commitString := "<!-- COMMIT: "
		off := strings.Index(html, commitString)
		closingComment := strings.Index(html, " -->")
		if closingComment > 0 {
			log.WithFields(log.Fields{
				"start":  off + len(commitString),
				"end":    closingComment,
				"string": html[off+len(commitString) : closingComment],
			}).Debug("match")
			return html[off+len(commitString) : closingComment], nil
		}
	}
	return "Unknown", nil
}
