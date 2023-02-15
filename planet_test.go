package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	pp, err := Load()
	require.NoError(t, err)
	for _, planet := range pp.Planets {
		for _, feedURL := range planet.Feeds {
			feedURL := feedURL
			t.Run(feedURL, func(t *testing.T) {
				t.Parallel()

				parser := gofeed.NewParser()
				feed, err := parser.ParseURL(feedURL)
				require.NoErrorf(t, err, "feed=%s", feedURL)
				require.NotEmpty(t, feed.Items, "feed=%", feedURL)
			})
		}
	}
}

func TestToRSS(t *testing.T) {
	pp := &PlanetPlanet{
		Planets: []Planet{
			{
				Title:       "title",
				Description: "description",
				HtmlLink:    "https://google.com",
				Link:        "https://google.com/index.xml",
				Output:      "google.xml",
				Feeds:       []string{},
			},
		},
	}
	items := []*gofeed.Item{}

	err := pp.ToRSS(items, &pp.Planets[0])
	require.NoError(t, err)

	f, err := os.Open(pp.Planets[0].Output)
	require.NoError(t, err)
	defer f.Close()
	feed, err := gofeed.NewParser().Parse(f)
	require.NoError(t, err)

	require.Equal(t, pp.Planets[0].HtmlLink, feed.Link)
}

func TestPlanetPlanet_GenerateIndex(t *testing.T) {
	type args struct {
		Author  string
		Email   string
		Planets []Planet
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default", args{Author: "author", Email: "email", Planets: []Planet{
			{
				Title:       "Title",
				Description: "Description",
				Link:        "http://feed.link",
				Feeds:       []string{"feed1", "feed2"},
			},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PlanetPlanet{
				Author:  tt.args.Author,
				Email:   tt.args.Email,
				Planets: tt.args.Planets,
			}
			if err := p.GenerateIndex(); (err != nil) != tt.wantErr {
				t.Errorf("PlanetPlanet.Index() error = %v, wantErr %v", err, tt.wantErr)
			}

			buf, err := os.ReadFile("index.html")
			require.NoError(t, err)

			for _, planet := range p.Planets {
				require.Contains(t, string(buf), fmt.Sprintf(`href="%s"`, planet.Link))
			}
		})
	}
}
