package main

import (
	"errors"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// A function that will be executed after a notification arrived it takes tht operation and the path
type EventAction func(fsnotify.Op, string) error

// Detect any operation happening on a specified path and reply executing an EventAction
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

// It manages the incoming events on the article folder
var ManageContents EventAction = func(eventType fsnotify.Op, filePath string) (err error) {
	var article *Article

	if eventType&fsnotify.Write != fsnotify.Write {
		// If article is deleted then move the html file into BACKUPS_FOLDER and delte from cache
		if eventType&fsnotify.Remove == fsnotify.Remove {
			fileName := strings.TrimSuffix(path.Base(strings.ReplaceAll(filePath, "\\", "/")), path.Ext(filePath))
			os.Rename(path.Join(BLOG_FOLDER, fileName+".html"), path.Join(BACKUPS_FOLDER, fileName+".html"))
			Cache.Remove(fileName)
		}
		return
	}

	// Parsing the article
	article, err = ReadArticle(filePath)
	if err != nil {
		return
	}

	// Renaming the article with his unique identifier
	fileName := strings.TrimSuffix(path.Base(strings.ReplaceAll(filePath, "\\", "/")), path.Ext(filePath))
	if _, e := os.Stat(path.Join(BLOG_FOLDER, fileName+".html")); os.IsNotExist(e) {
		article.Sign("DazFather", "", "", false, nil)

		err = os.Rename(filePath, path.Join(ARTICLE_FOLDER, article.RelativeLink+".md"))
		if err != nil {
			return
		}
		// If article have already the identifier as name, just update the signature
	} else {
		// extract the original date
		var date time.Time
		date, err = extractDate(fileName)
		if err != nil {
			return
		}

		// Sign but without generate a new identifier and set up the date to the original
		article.Sign("DazFather", "", fileName, false, &date)
	}

	// Generate HTML file
	err = article.SaveAsHTML(BLOG_FOLDER, "article.tmpl")
	if err != nil {
		return
	}

	// Save into chache
	_, err = Cache.Save(article)

	return
}

// It manages the incoming events on the template folder
var UpdateTemplates EventAction = func(eventType fsnotify.Op, filePath string) (err error) {
	if eventType&fsnotify.Write == fsnotify.Write || eventType&fsnotify.Remove == fsnotify.Remove {
		LoadTemplates(path.Dir(strings.ReplaceAll(filePath, "\\", "/")))
	}
	return
}
