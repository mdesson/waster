package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"sort"
	"strings"
)

type Site struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

//go:embed templates/*
var templatesFolder embed.FS

func main() {
	// init logger
	logHandler := slog.NewJSONHandler(os.Stdout, nil)
	l := slog.New(logHandler)
	l.Info("initialized logger")

	// init sites
	sites, err := initSites()
	if err != nil {
		log.Fatal(err)
	}

	l.Info("initailized sites")

	// init templates
	templates := template.Must(template.New("").ParseFS(templatesFolder, "templates/*.html"))
	l.Info("initialized templates")

	// Register handlers
	http.HandleFunc("/", indexHandler(templates, &sites, l))
	http.HandleFunc("/begin", beginHandler(templates, &sites, l))
	http.HandleFunc("/save", saveHandler(&sites, l))
	http.HandleFunc("/delete", deleteHandler(&sites, l))
	l.Info("initialized handlers")

	// Start server
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func initSites() ([]Site, error) {
	var sites []Site

	sitesBytes, err := os.ReadFile("sites.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(sitesBytes, &sites); err != nil {
		return nil, err
	}

	sort.Slice(sites, func(i, j int) bool {
		return strings.ToLower(sites[i].Title)[0] < strings.ToLower(sites[j].Title)[0]
	})

	return sites, nil
}

func overwriteSites(sites []Site) error {
	// write back to json to survive restarts
	f, err := os.OpenFile("sites.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer f.Close()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(sites, "", "\t")
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}
	return nil
}

// indexHandler shows the main page
func indexHandler(t *template.Template, sites *[]Site, l *slog.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// set data
		data := map[string]any{
			"sites": sites,
		}

		// render the page
		if err := t.ExecuteTemplate(w, "index.html", data); err != nil {
			l.Error("error loading index page", "error", err)
		}
	}
}

func beginHandler(t *template.Template, sites *[]Site, l *slog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// pick a site
		i := rand.IntN(len(*sites))
		site := (*sites)[i]

		l.Info("redirecting", "name", site.Title, "url", site.URL)

		// set cache-busting headers
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// render the page
		if err := t.ExecuteTemplate(w, "begin.html", site); err != nil {
			l.Error("error loading begin page", "error", err)
		}
	}
}

func saveHandler(sites *[]Site, l *slog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract site data and validate it's non-empty
		title := r.FormValue("name")
		url := r.FormValue("url")

		if title == "" || url == "" {
			w.WriteHeader(http.StatusBadRequest)
			l.Error("received bad add request", "title", title, "url", url)
			return
		}

		// save changes
		site := Site{Title: title, URL: url}
		updatedSites := append(*sites, site)

		if err := overwriteSites(updatedSites); err != nil {
			l.Error("error writing sites during add operation", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s, err := initSites()
		if err != nil {
			l.Error("error reinitializing sites after write back", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		*sites = s

		w.WriteHeader(http.StatusCreated)
	}
}

func deleteHandler(sites *[]Site, l *slog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract site data and validate it's non-empty
		url := r.FormValue("url")
		if url == "" {
			http.Error(w, "missing url", http.StatusBadRequest)
			return
		}

		// delete the item
		deleteIndex := -1
		for i, s := range *sites {
			if s.URL == url {
				deleteIndex = i
				break
			}
		}

		if deleteIndex == -1 {
			l.Error("bad request, site not found", "url", url)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updatedSites := append((*sites)[:deleteIndex], (*sites)[deleteIndex+1:]...)

		// write to disk and make alphabetical
		if err := overwriteSites(updatedSites); err != nil {
			l.Error("error writing sites during delete operation", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s, err := initSites()
		if err != nil {
			l.Error("error reinitializing sites after delete", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// update pointer
		*sites = s

		w.WriteHeader(http.StatusCreated)
	}
}
