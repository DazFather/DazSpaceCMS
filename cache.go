package main

import (
	"errors"
	"time"
)

// TODO: make theese attributes private
type Savings struct {
	SavedArticles map[string]*Article
	SavedSnippets map[string]*Snippet
	Timeline      []string
}

type PhantomSnippet struct {
	Link  string
	Title string
	Date  time.Time
}

const MEMORY_CAP = 10

var Cache = Savings{
	SavedArticles: make(map[string]*Article),
	SavedSnippets: make(map[string]*Snippet),
}

func (memory *Savings) trimOldest() {
	var (
		max  = time.Date(0, time.September, 1, 0, 0, 0, 0, time.UTC)
		link = ""
	)

	if len(memory.SavedSnippets) <= 0 {
		return
	}

	for currentLink, snippet := range memory.SavedSnippets {
		if max.Before(snippet.Date) {
			max = snippet.Date
			link = currentLink
		}
	}
	if link == "" {
		return
	}

	delete(memory.SavedSnippets, link)
	delete(memory.SavedArticles, link)
	memory.deleteFromTimeline(link)
}

func (memory *Savings) deleteFromTimeline(link string) (deleted bool) {
	// TODO: Upgrade speed using a bisection search (?)
	for i, current := range memory.Timeline {
		if current == link {
			memory.Timeline = append(memory.Timeline[:i], memory.Timeline[i+i:]...)
			deleted = true
			break
		}
	}

	return
}

func (memory *Savings) addToTimeline(link string, date time.Time) {
	var (
		last = memory.len() - 1
		ind  = last
	)

	for _, current := range memory.Timeline {
		if current == link {
			return
		}
		if ind >= 0 && date.Before(memory.SavedSnippets[memory.Timeline[ind]].Date) {
			ind--
		}
	}

	if ind == last {
		memory.Timeline = append(memory.Timeline, link)
		return
	}
	ind++

	memory.Timeline = append(memory.Timeline[:ind+1], memory.Timeline[ind:]...)
	memory.Timeline[ind] = link
}

func (memory *Savings) len() int {
	return len(memory.Timeline)
}

func (memory *Savings) Save(a *Article) (link string, err error) {
	link = a.RelativeLink
	if link == "" {
		err = errors.New("Missing relative link")
		return
	}

	for memory.len() >= MEMORY_CAP {
		memory.trimOldest()
	}

	snippet := a.Extract()
	memory.SavedArticles[link] = a
	memory.SavedSnippets[link] = &snippet
	memory.addToTimeline(link, snippet.Date)

	return
}

func (memory *Savings) SavePhantom(s Snippet) (link string, err error) {
	s.Abstract = "Preview not avaiable for this article"
	s.Title = extractTitle(s.Link)
	link = s.Link
	if link == "" {
		err = errors.New("Missing relative link")
		return
	}

	for memory.len() >= MEMORY_CAP {
		memory.trimOldest()
	}

	memory.SavedSnippets[link] = &s
	memory.addToTimeline(link, s.Date)

	return
}

func (memory *Savings) Remove(link string) (art *Article) {
	art = Cache.SavedArticles[link]

	delete(memory.SavedArticles, link)
	delete(memory.SavedSnippets, link)
	memory.deleteFromTimeline(link)

	return
}

func (memory *Savings) GetSnippets() (Collection []Snippet) {
	if memory.len() <= 0 {
		return
	}

	for _, identifier := range memory.Timeline {
		Collection = append(Collection, *memory.SavedSnippets[identifier])
	}
	return
}
