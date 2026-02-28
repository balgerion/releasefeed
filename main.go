package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"releasefeed/internal"
)

func main() {
	cfg, err := internal.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	st, err := internal.NewState("state.json")
	if err != nil {
		log.Fatal(err)
	}

	cache := internal.NewCache()
	app := internal.NewApprise(cfg.Apprise.URL)

	for _, f := range cfg.Feeds {
		f := f
		interval, err := time.ParseDuration(f.FetchInterval)
		if err != nil {
			log.Fatalf("invalid interval for feed %s: %v", f.Name, err)
		}

		poll := func() {
			releases, err := internal.Fetch(f.Name, f.URL)
			if err != nil {
				log.Printf("fetch error %s: %v", f.Name, err)
				return
			}
			normalized := make([]internal.Release, 0, len(releases))
			for _, r := range releases {
				r = internal.Normalize(r, f)
				normalized = append(normalized, r)
				if !st.Seen(r.ID) {
					if internal.HasKeyword(r.Title+" "+r.Content, f.AlertKeywords) {
						if err := app.Notify(r, f.AppriseTags); err != nil {
							log.Printf("apprise error %s: %v", f.Name, err)
						}
					}
					if err := st.Mark(r.ID); err != nil {
						log.Printf("state error: %v", err)
					}
				}
			}
			cache.Set(f.Name, normalized)
		}

		poll()

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			for range t.C {
				poll()
			}
		}()
	}

	http.HandleFunc("/feed", internal.FeedHandler(cache))

	go func() {
		log.Printf("listening on :%s", cfg.Server.Port)
		if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
