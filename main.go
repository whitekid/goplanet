package main

import (
	"context"
	"fmt"
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
		RunE:  func(cmd *cobra.Command, args []string) error { return updateRSS(cmd.Context()) },
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
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
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

func updateRSS(ctx context.Context) error {
	pp, err := Load()
	if err != nil {
		return err
	}

	for _, p := range pp.Planets {
		if err := pp.ToRSS(p.Load(ctx), &p); err != nil {
			return err
		}
	}

	return pp.Index()
}
