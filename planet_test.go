package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlanetPlanet_Index(t *testing.T) {
	type fields struct {
		Author  string
		Email   string
		Planets []Planet
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"default", fields{Author: "author", Email: "email", Planets: []Planet{
			Planet{
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
				Author:  tt.fields.Author,
				Email:   tt.fields.Email,
				Planets: tt.fields.Planets,
			}
			if err := p.Index(); (err != nil) != tt.wantErr {
				t.Errorf("PlanetPlanet.Index() error = %v, wantErr %v", err, tt.wantErr)
			}

			f, err := os.Open("index.html")
			require.NoError(t, err)
			defer f.Close()

			buf := bytes.Buffer{}
			_, err = io.Copy(&buf, f)
			require.NoError(t, err)

			for _, planet := range p.Planets {
				require.Contains(t, buf.String(), fmt.Sprintf(`href="%s"`, planet.Link))
			}
		})
	}
}
