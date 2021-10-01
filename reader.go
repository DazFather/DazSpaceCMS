package main

import (
	"bufio"
	"errors"
	"html/template"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/russross/blackfriday/v2"
)

func parse(text string) (parsed string) {
	var content = blackfriday.Run([]byte(text))

	content = regexp.MustCompile(`\s*</?p>\s*`).ReplaceAll(content, nil)
	return string(content)
}

// Scan of the title, cover art and other attached resources
func scanHeader(scanner *bufio.Scanner, art *Article) (err error) {
	var (
		cover                   string
		stylesName, scriptsName []string
		headerLink              = regexp.MustCompile(`!\[[A-Z]+\]\([\w\./:%&\?!=\- \\]+\)`)
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		err = scanner.Err()
		if err != nil {
			return
		}

		switch true {
		case line == "":
			continue

		case strings.HasPrefix(line, "#"):
			art.Title = strings.TrimSpace(line[1:])
			art.Clip(cover, stylesName, scriptsName)
			return

		case headerLink.Match([]byte(line)):
			link := strings.Split(line[2:len(line)-1], "](")
			switch link[0] {
			case "COVER":
				cover = link[1]
			case "SCRIPT":
				scriptsName = append(scriptsName, link[1])
			case "STYLE":
				stylesName = append(stylesName, link[1])
			}

		default:
			log.Println("line [" + line + "] is not a title")
			return errors.New("Missing title")
		}
	}

	return errors.New("Missing title")
}

func ReadArticle(filename string) (art *Article, err error) {
	var (
		file         *os.File
		scanner      *bufio.Scanner
		isNewChapter bool
		newArt       Article
		nCap         = -1
	)

	// Open the file creating to scan it
	file, err = os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	scanner = bufio.NewScanner(file)

	// Scan the header
	err = scanHeader(scanner, &newArt)
	if err != nil {
		return
	}

	// Scan the rest of the file
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check if line is blank
		if line == "" {
			if nCap > -1 && newArt.Chapters[nCap].Content != "" {
				newArt.Chapters[nCap].Content += "\n"
			}
			continue
		}

		// Check if is a new Chapter
		isNewChapter, err = regexp.Match(`^##[[:print:]]*`, []byte(line))
		if err != nil {
			return
		}
		if isNewChapter {
			line = parse(line[2:])
			newArt.Chapters = append(newArt.Chapters, Chapter{Name: line})
			nCap++
			continue
		}

		// Add content to Chapter
		if nCap == -1 {
			return nil, errors.New("paragraph without name")
		}
		newArt.Chapters[nCap].Content += template.HTML(parse(line) + "\n")
	}

	// Check for errors in the scanner
	err = scanner.Err()
	if err != nil {
		return
	}

	// Refining each content
	rgxEndln := regexp.MustCompile(`\r?\n`)
	for i, current := range newArt.Chapters {
		// Cleaning excess of '\n'
		raw := strings.TrimSpace(string(current.Content))
		// Dividing text in "<p>"
		raw = rgxEndln.ReplaceAllString(raw, "</p>\n<p>")
		newArt.Chapters[i].Content = template.HTML("<p>" + raw + "</p>")
	}

	// Saving and return
	art = &newArt
	return
}

// It manage the incoming events on the article folder
var ManageContents EventAction = func(eventType fsnotify.Op, filePath string) (err error) {
	var article *Article

	if eventType&fsnotify.Write != fsnotify.Write {
		// If article is deleted then move the html file into BACKUPS_FOLDER and delte from cache
		if eventType&fsnotify.Remove == fsnotify.Remove {
			fileName := strings.TrimSuffix(path.Base(strings.ReplaceAll(filePath, "\\", "/")), path.Ext(filePath))
			os.Rename(path.Join(BLOG_FOLDER, fileName+".html"), path.Join(BACKUPS_FOLDER, fileName+".html"))
			RemoveFromCache(fileName)
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
		article.Sign("DazFather", "", false)

		err = os.Rename(filePath, path.Join(ARTICLE_FOLDER, article.RelativeLink+".md"))
		if err != nil {
			return
		}
		// If article have already the identifier as name, just update the signature
	} else {
		// We can't use Sign or else a new identifier will be generated
		article.Date = strings.Join(strings.Split(fileName, "-")[1:4], " ")
		article.Author = "DazFather"
		article.AuthorLink = ""
		// As RelativeLink (identifier) we put its own fileName without extentions
		article.RelativeLink = fileName
	}

	// Generate HTML file
	err = article.SaveAsHTML(BLOG_FOLDER, "article.tmpl")
	if err != nil {
		return
	}

	// Save into chache
	_, err = article.SaveIntoCache()

	return
}

var UpdateTemplates EventAction = func(eventType fsnotify.Op, filePath string) (err error) {
	if eventType&fsnotify.Write == fsnotify.Write || eventType&fsnotify.Remove == fsnotify.Remove {
		LoadTemplates(path.Dir(strings.ReplaceAll(filePath, "\\", "/")))
	}
	return
}
