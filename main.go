package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{}

	root.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "update feeds",
		Run: func(cmd *cobra.Command, args []string) {
			pp, err := Load()
			if err != nil {
				log.Fatalf("Fail to load config: %s", err)
			}

			for _, p := range pp.Planets {
				pp.ToRSS(p.Load(), &p)
			}

			pp.Index()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "find url",
		Short: "find rss feed url",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(args[0])
			if err != nil {
				return err
			}

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				return err
			}

			doc.Find("link").Each(func(index int, s *goquery.Selection) {
				if val, ok := s.Attr("rel"); !ok || val != "alternate" {
					return
				}

				if val, ok := s.Attr("type"); !ok || val != "application/rss+xml" {
					return
				}

				if link, ok := s.Attr("href"); ok && len(link) > 0 {
					fmt.Printf("%s\n", link)
				}
			})
			return nil
		},
	})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
