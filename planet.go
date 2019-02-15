package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	ini "gopkg.in/ini.v1"
)

// PlanetPlanet ...
type PlanetPlanet struct {
	Author  string
	Email   string
	Planets []Planet
}

// Planet ...
type Planet struct {
	Title       string
	Description string
	Link        string
	Output      string
	Feeds       []string
}

// Load ...
func Load() (*PlanetPlanet, error) {
	cfg, err := ini.ShadowLoad("feeds.ini")
	if err != nil {
		return nil, err
	}

	p := &PlanetPlanet{}
	p.Author = cfg.Section("").Key("author").String()
	p.Email = cfg.Section("").Key("email").String()

	for _, section := range cfg.Sections() {
		if !strings.HasPrefix(section.Name(), "planet:") {
			continue
		}

		p.Planets = append(p.Planets,
			Planet{
				Title:       section.Name()[7:],
				Description: section.Key("description").String(),
				Output:      section.Key("output").String(),
				Link:        section.Key("link").String(),
				Feeds:       section.Key("feed").ValueWithShadows(),
			})
	}

	return p, nil
}

// ToRSS ...
func (p *PlanetPlanet) ToRSS(items []*gofeed.Item, planet *Planet) error {
	feed := &feeds.Feed{
		Title:       planet.Title,
		Link:        &feeds.Link{Href: planet.Link},
		Description: planet.Description,
		Author:      &feeds.Author{Name: p.Author, Email: p.Email},
		Created:     time.Now(),
		Items:       make([]*feeds.Item, len(items)),
	}

	for i, item := range items {
		created := item.PublishedParsed
		if created == nil {
			created = &time.Time{}
		}

		updated := item.UpdatedParsed
		if updated == nil {
			updated = &time.Time{}
		}

		author := &feeds.Author{}
		if item.Author != nil {

			author.Email = item.Author.Email
		}

		feed.Items[i] = &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Id:          item.Link,
			Created:     *created,
			Updated:     *updated,
			Description: item.Description,
			Content:     item.Content,
			Author:      author,
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Fatalf("Fail to generate rss: %s", err)
	}

	f, err := os.OpenFile(planet.Output, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("fail to open file %s: %s", planet.Output, err)
	}
	defer f.Close()
	f.WriteString(rss)

	return nil
}

// Index generate index file
func (p *PlanetPlanet) Index() error {
	tpl, err := pongo2.FromFile("index.tmpl")
	if err != nil {
		return err
	}

	context := map[string]interface{}{
		"planetplanet": p,
	}

	out, err := tpl.Execute(context)
	if err != nil {
		return err
	}

	f, err := os.OpenFile("index.html", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(out)

	return nil
}

// Load ...
func (p *Planet) Load() []*gofeed.Item {
	var wg sync.WaitGroup
	feedC := make(chan string, 30)
	resultC := make(chan []*gofeed.Item, 30)

	// start worker
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(ID int) {
			defer wg.Done()
			for feedURL := range feedC {
				t := time.Now()
				parser := gofeed.NewParser()
				feed, err := parser.ParseURL(feedURL)
				if err != nil {
					log.Printf("Fail to parse feed: %s", p.Title)
					return
				}

				log.Printf("[%d] %s has %d items. %s",
					ID, feedURL, len(feed.Items), time.Since(t))
				resultC <- feed.Items
			}
		}(i)
	}

	// close result channel when worker done
	go func() {
		wg.Wait()
		close(resultC)
	}()

	// feed
	for _, u := range p.Feeds {
		feedC <- u
	}
	close(feedC)

	items := []*gofeed.Item{}
	for r := range resultC {
		items = append(items, r...)
	}

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

	return items
}
