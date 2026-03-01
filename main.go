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
	cfg, err := internal.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	st, err := internal.NewState("config/state.json")
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
			log.Printf("[%s] fetching", f.Name)
			releases, err := internal.Fetch(f.Name, f.URL)
			if err != nil {
				log.Printf("[%s] fetch error: %v", f.Name, err)
				return
			}
			log.Printf("[%s] got %d releases", f.Name, len(releases))
			normalized := make([]internal.Release, 0, len(releases))
			for _, r := range releases {
				r = internal.Normalize(r, f)
				r.Image = f.Image
				normalized = append(normalized, r)
				if !st.Seen(r.ID) {
					if internal.HasKeyword(r.Title+" "+r.Content, f.AlertKeywords) {
						log.Printf("[%s] alert: %s", f.Name, r.Title)
						if err := app.Notify(r, f.AppriseTags); err != nil {
							log.Printf("[%s] apprise error: %v", f.Name, err)
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

	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("config/images"))))
	http.HandleFunc("/feed", internal.FeedHandler(cache, cfg.Server.BaseURL))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

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