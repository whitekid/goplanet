package main

import "log"

func main() {
	pp, err := Load()
	if err != nil {
		log.Fatalf("Fail to load config: %s", err)
	}

	for _, p := range pp.Planets {
		pp.ToRSS(p.Load(), &p)
	}
}
