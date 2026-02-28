package internal

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type Release struct {
	ID          string
	FeedName    string
	Title       string
	Content     string
	URL         string
	PublishedAt time.Time
    Image       string
}

func Fetch(feedName, url string) ([]Release, error) {
	fp := gofeed.NewParser()
	f, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}
	releases := make([]Release, 0, len(f.Items))
	for _, item := range f.Items {
		r := Release{
			ID:       item.GUID,
			FeedName: feedName,
			Title:    item.Title,
			URL:      item.Link,
		}
		if item.Content != "" {
			r.Content = item.Content
		} else {
			r.Content = item.Description
		}
		if item.PublishedParsed != nil {
			r.PublishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			r.PublishedAt = *item.UpdatedParsed
		}
		releases = append(releases, r)
	}
	return releases, nil
}
