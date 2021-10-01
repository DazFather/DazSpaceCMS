package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Chapter struct {
	Name    string
	Content template.HTML
}

type Article struct {
	RelativeLink    string
	Title           string
	Date            string
	Author          string
	AuthorLink      string
	Chapters        []Chapter
	AttachedCover   string
	AttachedScripts []string
	AttachedStyles  []string
	Unlisted        bool
}

type Snippet struct {
	Title    string
	Abstract string
	Cover    string
	Link     string
}

var Cache = make(map[string]*Article)

func SelectFromChache(snip string) (art *Article) {
	return Cache[snip]
}

func (a *Article) SaveIntoCache() (link string, err error) {
	link = a.RelativeLink
	if link == "" {
		err = errors.New("Missing relative link")
		return
	}
	Cache[link] = a

	return
}

func RemoveFromCache(snip string) (art *Article) {
	art = Cache[snip]
	delete(Cache, snip)
	return
}

func (art *Article) SaveAsHTML(folderPath, templateName string) error {
	if art.RelativeLink == "" {
		return errors.New("Missing RelativeLink")
	}

	file, err := os.Create(path.Join(folderPath, art.RelativeLink+".html"))
	defer file.Close()
	if err != nil {
		return err
	}

	return Compose(file, templateName, art)
}

func (a *Article) GenLink() string {
	if a.Title == "" || a.Date == "" {
		return ""
	}

	var (
		title = url.QueryEscape(a.Title + "-" + strings.ReplaceAll(a.Date, " ", "-"))
		i     = 1
	)

	a.RelativeLink = title
	for Cache[a.RelativeLink] != nil {
		a.RelativeLink = fmt.Sprint(title, "-", i)
		i++
	}
	return a.RelativeLink
}

func (a *Article) Clip(cover string, stylesName, scriptsName []string) *Article {
	// Clipping Cover
	if cover != "" {
		switch path.Dir(cover) {
		case ".", "/":
			a.AttachedCover = "/" + path.Join(IMAGES_FOLDER, cover)
		default:
			a.AttachedCover = cover
		}
	}

	// Clipping Styles
	for i := range stylesName {
		dir := path.Dir(stylesName[i])
		if dir == "." || dir == "/" {
			stylesName[i] = "/" + path.Join(STYLES_FOLDER, stylesName[i])
		}
	}
	a.AttachedStyles = stylesName

	// Clipping Scripts
	for i := range scriptsName {
		dir := path.Dir(scriptsName[i])
		if dir == "." || dir == "/" {
			scriptsName[i] = path.Join(SCRIPTS_FOLDER, scriptsName[i])
		}
	}
	a.AttachedScripts = scriptsName

	return a
}

func (a *Article) Sign(author, link string, hidden bool) *Article {
	a.Date = fmt.Sprint(time.Now().Date())
	a.Author = author
	a.AuthorLink = link
	a.Unlisted = hidden
	a.GenLink()
	return a
}

func (a *Article) Extract() Snippet {
	var (
		htmltag = regexp.MustCompile(`<.+?>`)
		text    = string(a.Chapters[0].Content)
	)

	return Snippet{
		Title:    a.Title,
		Abstract: htmltag.ReplaceAllString(text, "") + "...",
		Cover:    a.AttachedCover,
		Link:     a.RelativeLink,
	}
}

func GenLastArticles() (Collection []Snippet) {
	for _, article := range Cache {
		if !article.Unlisted {
			Collection = append(Collection, article.Extract())
		}
	}

	return
}
