package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var views = template.Must(template.New("").
	Funcs(template.FuncMap{"isCurrent": isCurrent}).
	ParseGlob("templates/*.html"))

var latest = map[string]string{}

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
	// MEFE, aka https://github.com/unee-t/frontend/ aka Meteor
	mefeVersions, err := getVersion([]Service{
		{Site: "https://case.dev.unee-t.com"},
		{Site: "https://case.demo.unee-t.com"},
		{Site: "https://case.unee-t.com"},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// dash, aka https://github.com/unee-t/frontend/ aka Bugzilla
	dashVersions, err := getVersion([]Service{
		{Site: "https://dashboard.dev.unee-t.com"},
		{Site: "https://dashboard.demo.unee-t.com"},
		{Site: "https://dashboard.unee-t.com"},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = views.ExecuteTemplate(w, "index.html", struct {
		Frontend []Service
		Bugzilla []Service
	}{
		mefeVersions,
		dashVersions,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getVersion(check []Service) (version []Service, err error) {
	for i, v := range check {
		resp, err := http.Get(v.Site)
		if err != nil {
			return check, err
		}
		defer resp.Body.Close()
		if strings.Index(v.Site, "case") > 0 {
			check[i].Version, err = parseCommitComment(resp.Body)
			if err != nil {
				return check, err
			}
			log.WithField("site", check[i]).Info("case")
		}
		if strings.Index(v.Site, "dash") > 0 {
			check[i].Version, err = parseHTMLspan(resp.Body)
			if err != nil {
				return check, err
			}
			log.WithField("site", check[i]).Info("dash")
		}

	}
	return check, nil
}

func parseHTMLspan(input io.Reader) (version string, err error) {
	// <span id="information" class="header_addl_info col-sm-3">e88ec7fdc</span>
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		html := scanner.Text()
		commitString := `<span id="information" class="header_addl_info col-sm-3">`
		off := strings.Index(html, commitString)
		log.WithFields(log.Fields{
			"start": off,
		}).Debug("first")
		if off >= 0 {
			closingComment := strings.Index(html[off:], "</span>")
			if closingComment > 0 {
				log.WithFields(log.Fields{
					"start": off,
					"end":   off + closingComment,
				}).Debug("match")
				return html[off+len(commitString) : off+closingComment], nil
			}
		}
	}
	return "", nil
}

func parseCommitComment(input io.Reader) (version string, err error) {
	// Version string is actually a commit id, e.g. "ae5b321" from:
	// <!-- COMMIT: ae5b321 -->
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		html := scanner.Text()
		commitString := "<!-- COMMIT: "
		off := strings.Index(html, commitString)
		log.WithFields(log.Fields{
			"start": off,
		}).Debug("first")
		if off >= 0 {
			closingComment := strings.Index(html[off:], " -->")
			if closingComment > 0 {
				log.WithFields(log.Fields{
					"start": off,
					"end":   off + closingComment,
				}).Debug("match")
				match := html[off+len(commitString) : off+closingComment]
				log.WithField("match", match).Debug("found")
				return strings.Split(match, " ")[0], nil
			}
		}
	}
	return "", nil
}

func isCurrent(url string, hash string) bool {
	if latest[url] == "" {
		remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
			URLs: []string{url},
		})
		log.WithField("url", url).Info("listing")
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			log.WithError(err).WithField("url", url).Error("not a git repo")
			return false
		}
		for _, r := range refs {
			if r.Name() == "refs/heads/master" {
				latest[url] = fmt.Sprintf("%s", r.Hash())
			}
		}
	} else {
		log.WithField("url", url).Info("known")
	}
	if strings.HasPrefix(latest[url], hash) {
		return true
	}
	return false
}
