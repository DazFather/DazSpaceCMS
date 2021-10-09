package main

import (
	"errors"
	"html/template"
	"time"
)

// Architecture of the Cache memory
type Savings struct {
	savedArticles  map[string]*Article // Articles saved on memory
	savedSnippets  map[string]*Snippet // Article's snippets saved on memory
	timeline       []string            // Orderd (by date) list of links that can be used to select to articles or snippets
	SavedTemplates *template.Template  // Html templates used on templates.go
}

// Max number of snippets and articles that will be saved on memory
const MEMORY_CAP = 10

// Cache memory
var Cache = Savings{
	savedArticles: make(map[string]*Article),
	savedSnippets: make(map[string]*Snippet),
}

// Get the number of links saved on cache
func (memory *Savings) Len() int {
	return len(memory.timeline)
}

// Save an article (and relative snippet) in the memory
func (memory *Savings) Save(a *Article) (link string, err error) {
	link = a.RelativeLink
	if link == "" {
		err = errors.New("Missing relative link")
		return
	}

	for memory.Len() >= MEMORY_CAP {
		memory.trimOldest()
	}

	snippet := a.Extract()
	memory.savedArticles[link] = a
	memory.savedSnippets[link] = &snippet
	memory.addTotimeline(link, snippet.Date)

	return
}

/* Save a phantom snippet in the memory
 * phantom snippet are Snippet that are related to an article that is no longer
 * avaiable as markdown but it still exist as html file
 */
func (memory *Savings) SavePhantom(s Snippet) (link string, err error) {
	s.Abstract = "Preview not avaiable for this article"
	s.Title = extractTitle(s.Link)
	link = s.Link
	if link == "" {
		err = errors.New("Missing relative link")
		return
	}

	for memory.Len() >= MEMORY_CAP {
		memory.trimOldest()
	}

	memory.savedSnippets[link] = &s
	memory.addTotimeline(link, s.Date)

	return
}

// Remove an article (and relative snippet) from the memory and return it
func (memory *Savings) Remove(link string) (art *Article) {
	art = memory.savedArticles[link]

	delete(memory.savedArticles, link)
	delete(memory.savedSnippets, link)
	memory.deleteFromtimeline(link)
	return
}

// Return an article form the memory, nil if there is not
func (memory *Savings) SelectArticle(link string) *Article {
	return memory.savedArticles[link]
}

// Return an snippet form the memory, nil if there is not
func (memory *Savings) SelectSnippet(link string) *Snippet {
	return memory.savedSnippets[link]
}

// Return the oldest article/snippet link
func (memory *Savings) Oldest() string {
	if memory.Len() <= 0 {
		return ""
	}
	return memory.timeline[0]
}

// Return the youngest article/snippet link
func (memory *Savings) Youngest() string {
	if memory.Len() <= 0 {
		return ""
	}
	return memory.timeline[memory.Len()-1]
}

/* Generate an orderd (by date) collection of snippets that are saved in memory
 * The number of them, is memory.Len() and it can be as max MEMORY_CAP
 */
func (memory *Savings) GenSnippets() (Collection []Snippet) {
	if memory.Len() <= 0 {
		return
	}

	for _, identifier := range memory.timeline {
		Collection = append(Collection, *memory.savedSnippets[identifier])
	}
	return
}

// Remove the oldest article and snippet form the memory
func (memory *Savings) trimOldest() {
	var link = memory.Oldest()

	if link == "" {
		return
	}

	delete(memory.savedSnippets, link)
	delete(memory.savedArticles, link)
	memory.timeline = memory.timeline[:memory.Len()-1]
}

// Delete a link to article and snippet from the timeline
func (memory *Savings) deleteFromtimeline(link string) (deleted bool) {
	// TODO: Upgrade speed using a bisection search (?)
	for i, current := range memory.timeline {
		if current == link {
			memory.timeline = append(memory.timeline[:i], memory.timeline[i+i:]...)
			deleted = true
			break
		}
	}

	return
}

// Add a link to article and snippet to the timeline
func (memory *Savings) addTotimeline(link string, date time.Time) {
	var ind int

	for i, current := range memory.timeline {
		if current == link {
			return
		}
		if snip := memory.SelectSnippet(current); snip != nil && date.After(snip.Date) {
			ind = i
			break
		}
	}

	if ind == 0 {
		memory.timeline = append([]string{link}, memory.timeline...)
		return
	}

	memory.timeline = append(memory.timeline[:ind+1], memory.timeline[ind:]...)
	memory.timeline[ind] = link
}
