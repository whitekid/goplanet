package main

import (
	"context"
	"os"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/whitekid/goxp"
	"github.com/whitekid/goxp/log"
	"github.com/whitekid/iter"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// PlanetPlanet ...
type PlanetPlanet struct {
	Author  string   `yaml:"author"`
	Email   string   `yaml:"email"`
	Planets []Planet `yaml:"planet"`
}

// Planet ...
type Planet struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	HtmlLink    string   `yaml:"htmlLink"` // html link
	Link        string   `yaml:"link"`     // rss link
	Output      string   `yaml:"output"`
	Feeds       []string `yaml:"feeds"`
}

// Load ...
func Load() (*PlanetPlanet, error) {
	f, err := os.Open("feeds.yaml")
	if err != nil {
		return nil, err
	}

	p := &PlanetPlanet{}
	if err := yaml.NewDecoder(f).Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

// ToRSS ...
func (p *PlanetPlanet) ToRSS(items []*gofeed.Item, planet *Planet) error {
	if len(items) > 20 {
		items = items[:19]
	}

	feed := &feeds.Feed{
		Title:       planet.Title,
		Link:        &feeds.Link{Href: planet.HtmlLink},
		Description: planet.Description,
		Author:      &feeds.Author{Name: p.Author, Email: p.Email},
		Updated:     time.Now(),
	}

	feed.Items = iter.Map(iter.S(items), func(item *gofeed.Item) *feeds.Item {
		created := goxp.Ternary(item.PublishedParsed != nil, item.PublishedParsed, &time.Time{})
		updated := goxp.Ternary(item.UpdatedParsed != nil, item.UpdatedParsed, &time.Time{})

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
	}).Slice()

	rss, err := feed.ToRss()
	if err != nil {
		return errors.Wrap(err, "fail to generate rss")
	}

	f, err := os.OpenFile(planet.Output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "fail to open file %s", planet.Output)
	}
	defer f.Close()
	_, err = f.WriteString(rss)

	return err
}

// GenerateIndex generate index file
func (p *PlanetPlanet) GenerateIndex() error {
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
	eg, _ := errgroup.WithContext(ctx)
	iter.C(feedC).Each(func(feedURL string) {
		eg.Go(func() error {
			t := time.Now()
			parser := gofeed.NewParser()
			feed, err := parser.ParseURL(feedURL)
			if err != nil {
				log.Errorf("fail to parse feed: %s", feedURL)
				return err
			}

			log.Infof("%s has %d items. %s",
				feedURL, len(feed.Items), time.Since(t))
			resultC <- feed.Items
			return nil
		})
	})
	eg.Wait()
	close(resultC)

	items := []*gofeed.Item{}
	iter.C(resultC).Each(func(item []*gofeed.Item) { items = append(items, item...) })

	return iter.SortedFunc(iter.S(items), func(a, b *gofeed.Item) bool {
		timea := a.PublishedParsed
		if timea == nil {
			timea = a.UpdatedParsed
		}

		timeb := b.PublishedParsed
		if timeb == nil {
			timeb = b.UpdatedParsed
		}

		return timea.After(*timeb)
	}).Slice()
}
