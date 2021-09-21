package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var (
	ResourcesPath = "/" + RESOURCE_DIR + "/"
)

func check(funcName string, err error) {
	if err != nil {
		log.Fatal(funcName, " ", err)
	}
}

// Handle the resources serving them statically
var HandlerResources = http.StripPrefix(ResourcesPath,
	http.FileServer(http.Dir("."+ResourcesPath)),
)

// Handle the homepage (.../ or .../index.html)
func HandlerHome(w http.ResponseWriter, r *http.Request) {
	err := Compose(w, "home.tmpl", GenLastArticles())
	check("HandlerHome", err)
}

// Handle the 404 Error
func Handle404(w http.ResponseWriter, r *http.Request) {
	err := Compose(w, "404.tmpl", nil)
	check("Handle404", err)
}

// Handle the Blog (.../Blog or .../Blog/...)
func HandlerBlog(w http.ResponseWriter, r *http.Request) {
	var (
		content   []byte
		requested string
		err       error
	)

	// Get the unique identifier of the requested article
	_, requested = path.Split(r.URL.Path)
	if requested == "" {
		HandlerHome(w, r)
		return
	}
	requested = strings.TrimSuffix(requested, ".html")

	// Check if it's on cache
	if article := SelectFromChache(requested); article != nil {
		err = Compose(w, "article.tmpl", article)
		if err == nil {
			return
		}
		log.Println("HandlerBlog", "SelectFromChache", err)
		RemoveFromCache(requested)
	}

	// Check inside the contents
	content, err = os.ReadFile(path.Join(BLOG_FOLDER, requested+".html"))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}
		Handle404(w, r)
		return
	}
	w.Write(content)
}

type EventAction func(fsnotify.Op, string) error

func Detect(PathToWath string, action EventAction) (err error) {
	var watch *fsnotify.Watcher

	watch, err = fsnotify.NewWatcher()
	if err != nil {
		log.Println("Detect", err)
		return
	}
	defer watch.Close()

	err = watch.Add(PathToWath)
	if err != nil {
		return
	}

	for {
		select {
		case event, ok := <-watch.Events:
			if !ok {
				return errors.New("Internal error meanwhile watching events")
			}
			// TODO: better managment of chache and handling deleting of articles
			if e := action(event.Op, event.Name); e != nil {
				log.Println("Detect", e)
			}

		case e, ok := <-watch.Errors:
			if !ok {
				return errors.New("Internal error meanwhile watching errors")
			}
			return e
		}
	}
}

func main() {
	// Load all templates from folder
	LoadTemplates(TEMPLATE_FOLDER)

	// Manage changes in the articles (new / edited / deleted)
	go Detect(ARTICLE_FOLDER, ManageContents)

	// Manage changes in the templates
	go Detect(TEMPLATE_FOLDER, UpdateTemplates)

	// All the resources with static handles
	http.Handle(ResourcesPath, HandlerResources)

	// Homepage
	http.HandleFunc("/", HandlerHome)

	// Blog
	http.HandleFunc("/blog/", HandlerBlog)

	// Launch server
	http.ListenAndServe(":8080", nil)
}
