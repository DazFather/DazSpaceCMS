package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	Description     string
	RelatedArticles []Snippet
	Unlisted        bool
}

type Snippet struct {
	Title    string
	Abstract string
	Cover    string
	Link     string
	Date     time.Time
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
	for Cache.SelectArticle(a.RelativeLink) != nil {
		a.RelativeLink = fmt.Sprint(title, "-v", i)
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
			scriptsName[i] = "/" + path.Join(SCRIPTS_FOLDER, scriptsName[i])
		}
	}
	a.AttachedScripts = scriptsName

	return a
}

func (a *Article) Sign(author, authorLink, link string, hidden bool, date *time.Time) *Article {
	if date == nil {
		a.Date = fmt.Sprint(time.Now().Date())
	} else {
		a.Date = fmt.Sprint(date.Date())
	}
	a.Author = author
	a.AuthorLink = authorLink
	a.Unlisted = hidden
	if link == "" {
		a.GenLink()
	} else {
		a.RelativeLink = link
	}

	return a
}

func (a *Article) Snip() Snippet {
	sp := Cache.SelectSnippet(a.RelativeLink)
	if sp != nil {
		return *sp
	}

	return a.Extract()
}

func (a *Article) Extract() Snippet {
	return Snippet{
		Title:    a.Title,
		Abstract: a.Description,
		Cover:    a.AttachedCover,
		Link:     a.RelativeLink,
	}
}

func extractDate(identifier string) (date time.Time, err error) {
	const dateform = "2006-January-2"
	var (
		re         = regexp.MustCompile(`2[0-9]{3}-[a-zA-Z]+-\d{1,2}(-v\d+)?$`)
		stringDate = re.FindString(identifier)
	)

	if ind := strings.Index(stringDate, "-v"); ind != -1 {
		stringDate = stringDate[:ind]
	}

	return time.Parse(dateform, stringDate)
}

func extractTitle(identifier string) (title string) {
	var re = regexp.MustCompile(`-2[0-9]{3}-[a-zA-Z]+-\d{1,2}(-v\d+)?$`)

	indexes := re.FindStringIndex(identifier)
	if indexes == nil {
		return
	}

	title, _ = url.QueryUnescape(identifier[:indexes[0]])
	return
}

func GenLastArticles() (Collection []Snippet) {
	var CurrentDate = time.Now()

	// Reading the directory to grab the names of the asrticles
	var files, err = os.ReadDir(BLOG_FOLDER)
	if err != nil {
		return
	}

	// Genereting and selecting an ordered array of raw (with just Date and Link) Snippets
	for _, file := range files {
		// Extract the identifier and the date of the Snippet from the file.Name()
		link := strings.TrimSuffix(path.Base(file.Name()), filepath.Ext(file.Name()))
		date, _ := extractDate(link)
		var newSnippet = Snippet{Link: link, Date: date}
		// Check if the article is old less then 1 year (approx.)
		if CurrentDate.Sub(date).Hours()/24 < 365 {
			// Add it to the Collection
			Collection = insert(newSnippet, Collection)
			// If the array is too big cut the last snippet out
			if len(Collection) >= MEMORY_CAP {
				Collection = Collection[:MEMORY_CAP]
			}
		}
	}

	// Grab the rest of the infos (Title, Cover, Abstract) for each snippet in the collection
	for _, rawSnippet := range Collection {
		// Check if is on cache
		if Cache.SelectSnippet(rawSnippet.Link) != nil {
			continue
		}

		// Parse the original article
		art, err := ReadArticle(path.Join(ARTICLE_FOLDER, rawSnippet.Link+".md"))
		if err != nil {
			Cache.SavePhantom(rawSnippet)
			continue
		}
		// Cannot use art.Sign(...) because it will generate a new link
		art.Sign("DazFather", "", rawSnippet.Link, false, &rawSnippet.Date)

		// Save it on cache
		Cache.Save(art)
	}

	return
}

func insert(newValue Snippet, Collection []Snippet) (newCollection []Snippet) {
	var (
		last = len(Collection) - 1
		i    = last
	)

	for i >= 0 && newValue.Date.Before(Collection[i].Date) {
		i--
	}

	if i == last {
		return append(Collection, newValue)
	}
	i++
	newCollection = append(Collection[:i+1], Collection[i:]...)
	newCollection[i] = newValue

	return
}
