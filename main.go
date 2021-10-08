package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var ResourcesPath = "/" + RESOURCE_DIR + "/"

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
	err := Compose(w, "home.tmpl", Cache.GenSnippets())
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
	if article := Cache.SelectArticle(requested); article != nil {
		err = Compose(w, "article.tmpl", article)
		if err == nil {
			return
		}
		log.Println("HandlerBlog", "Cache.SelectArticle", err)
		Cache.Remove(requested)
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

// Load template and create missing folders
func InitializeServer() {
	// Loads all settings and paths
	LoadSettings()

	// Load all templates from folders
	LoadTemplates(TEMPLATE_FOLDER)

	// Recreate missing folders
	check("HealDirectories", HealDirectories())

	GenLastArticles()
}

func main() {
	// Initialize server loading templates
	InitializeServer()

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
