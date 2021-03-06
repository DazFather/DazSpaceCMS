package main

import (
	"bufio"
	"errors"
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// Transform a markdown text into HTML
func parse(text string) (parsed template.HTML) {
	var content = blackfriday.Run([]byte(text))
	return template.HTML(content)
}

// Scan of the title, cover art and other attached resources
func scanHeader(scanner *bufio.Scanner, art *Article) (err error) {
	var (
		cover                   string
		stylesName, scriptsName []string
		headerLink              = regexp.MustCompile(`^!\[[A-Z]+\](\([\w\./:%&\?!=\- \\]+\))+$`)
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
			art.Clip(cover, stylesName, scriptsName)
			return

		case headerLink.Match([]byte(line)):
			link := strings.Split(line[2:len(line)-1], "](")
			switch link[0] {
			case "COVER":
				cover = link[1]
			case "SCRIPT":
				scriptsName = append(scriptsName, strings.Split(link[1], ")(")...)
			case "STYLE":
				stylesName = append(stylesName, strings.Split(link[1], ")(")...)
			case "DESCRIPTION":
				art.Description = template.HTMLEscapeString(link[1])
			}

		default:
			log.Println("line [" + line + "] is not a title")
			return errors.New("Missing title")
		}
	}

	return errors.New("Missing title")
}

// Parse a valid markdown file and return a pointer to a new article (and an error)
func ReadArticle(filename string) (art *Article, err error) {
	var (
		file         *os.File
		scanner      *bufio.Scanner
		isNewChapter bool
		newArt       Article
		nCap         = -1
		paragraph    = regexp.MustCompile(`^\s*<p>|</p>\s*$|\s*<p>\s*</p>\s*`)
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
		line := scanner.Text()

		// Check if line is blank
		if strings.TrimSpace(line) == "" {
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
			line = strings.TrimSpace(line[2:])
			line = paragraph.ReplaceAllString(string(parse(line)), "")
			newArt.Chapters = append(newArt.Chapters, Chapter{Name: template.HTML(line)})
			nCap++
			continue
		}

		// Add content to Chapter
		if nCap == -1 {
			return nil, errors.New("paragraph without name")
		}
		newArt.Chapters[nCap].Content += template.HTML(line + "\n")
	}

	// Check for errors in the scanner
	err = scanner.Err()
	if err != nil {
		return
	}

	// Parsing each content
	for i, current := range newArt.Chapters {
		newArt.Chapters[i].Content = parse(string(current.Content))
	}

	// Generete a description if not present
	if newArt.Description == "" && len(newArt.Chapters) >= 1 {
		newArt.Description = string(newArt.Chapters[0].Content)
		newArt.Description = template.HTMLEscapeString(
			regexp.MustCompile(`<.+?>`).ReplaceAllString(newArt.Description, ""),
		)
	}
	newArt.Description = strings.TrimSpace(newArt.Description)
	if len(newArt.Description) > 80 {
		ind := strings.LastIndex(newArt.Description[:80], " ")
		newArt.Description = newArt.Description[:ind] + "..."
	}

	// Saving and return
	art = &newArt
	return
}
