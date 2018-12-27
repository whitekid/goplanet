package main

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	ini "gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("feeds.ini")
	if err != nil {
		panic(err)
	}

	items := []*gofeed.Item{}

	// load feeds
	parser := gofeed.NewParser()
	for _, section := range cfg.Sections() {
		url, err := url.Parse(section.Name())
		if err != nil {
			log.Printf("Fail to parse feed url: %s, %s", section.Name(), err)
			continue
		}

		if !(url.Scheme == "http" || url.Scheme == "https") {
			log.Printf("not a valid feed url: %s", section.Name())
			continue
		}

		feed, err := parser.ParseURL(section.Name())
		if err != nil {
			log.Printf("Fail to parse feed: %s", section.Name())
			continue
		}

		items = append(items, feed.Items...)
		// log.Printf("%d item, %s, %q", len(feed.Items), section.Name(), feed.Items)
	}

	// sort
	log.Printf("%d items was collected\n", len(items))
	sort.Slice(items, func(i, j int) bool {
		timei := items[i].PublishedParsed
		if timei == nil {
			timei = items[i].UpdatedParsed
		}
		timej := items[j].PublishedParsed
		if timej == nil {
			timej = items[j].UpdatedParsed
		}

		return timei.After(*timej)
	})

	// Generate feeds
	feed := &feeds.Feed{
		Title:       "Go Planet",
		Link:        &feeds.Link{Href: "http://goplanet.github.io"},
		Description: "Golang Blog Planet",
		Author:      &feeds.Author{Name: "Charlie.Choe", Email: "whitekid@gmail.com"},
		Created:     time.Now(),
		Items:       make([]*feeds.Item, len(items)),
	}
	for i, item := range items {
		tm := item.PublishedParsed
		if tm == nil {
			tm = item.UpdatedParsed
		}

		feed.Items[i] = &feeds.Item{
			Title:   item.Title,
			Link:    &feeds.Link{Href: item.Link},
			Created: *tm,
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Fatalf("Fail to generate rss: %s", err)
	}

	fmt.Printf("%s", rss)
}
