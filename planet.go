package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/whitekid/goxp"
	"github.com/whitekid/goxp/fx"
	"github.com/whitekid/goxp/log"
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

	sections := fx.Filter(cfg.Sections(), func(section *ini.Section) bool {
		return strings.HasPrefix(section.Name(), "planet:")
	})
	p.Planets = fx.Map(sections, func(section *ini.Section) Planet {
		return Planet{
			Title:       section.Name()[7:],
			Description: section.Key("description").String(),
			Output:      section.Key("output").String(),
			Link:        section.Key("link").String(),
			Feeds:       section.Key("feed").ValueWithShadows(),
		}
	})

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
	}

	feed.Items = fx.Map(items, func(item *gofeed.Item) *feeds.Item {
		created := fx.Ternary(item.PublishedParsed != nil, item.PublishedParsed, &time.Time{})
		updated := fx.Ternary(item.UpdatedParsed != nil, item.UpdatedParsed, &time.Time{})

		author := &feeds.Author{}
		if item.Author != nil {
			author.Email = item.Author.Email
		}

		return &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Id:          item.Link,
			Created:     *created,
			Updated:     *updated,
			Description: item.Description,
			Content:     item.Content,
			Author:      author,
		}
	})

	rss, err := feed.ToRss()
	if err != nil {
		return errors.Wrap(err, "fail to generate rss")
	}

	f, err := os.OpenFile(planet.Output, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "fail to open file %s", planet.Output)
	}
	defer f.Close()
	_, err = f.WriteString(rss)

	return err
}

// Index generate index file
func (p *PlanetPlanet) Index() error {
	tpl, err := pongo2.FromFile("index.tmpl")
	if err != nil {
		return err
	}

	out, err := tpl.Execute(pongo2.Context{"planetplanet": p})
	if err != nil {
		return err
	}

	f, err := os.OpenFile("index.html", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(out)
	return err
}

// Load ...
func (p *Planet) Load(ctx context.Context) []*gofeed.Item {
	feedC := make(chan string, 30)
	resultC := make(chan []*gofeed.Item, 30)

	// feed
	go func() {
		for _, u := range p.Feeds {
			feedC <- u
		}
		close(feedC)
	}()

	// start worker
	goxp.DoWithWorker(ctx, 5, func(i int) error {
		fx.IterChan(ctx, feedC, func(feedURL string) {
			t := time.Now()
			parser := gofeed.NewParser()
			feed, err := parser.ParseURL(feedURL)
			if err != nil {
				log.Infof("Fail to parse feed: %s", p.Title)
				return
			}

			log.Infof("[%d] %s has %d items. %s",
				i, feedURL, len(feed.Items), time.Since(t))
			resultC <- feed.Items
		})
		return nil
	})
	close(resultC)

	items := []*gofeed.Item{}
	fx.IterChan(ctx, resultC, func(item []*gofeed.Item) { items = append(items, item...) })

	items = fx.SortFunc(items, func(a, b *gofeed.Item) bool {
		timea := a.PublishedParsed
		if timea == nil {
			timea = a.UpdatedParsed
		}

		timeb := b.PublishedParsed
		if timeb == nil {
			timeb = b.UpdatedParsed
		}

		return timea.After(*timeb)
	})

	return items
}
