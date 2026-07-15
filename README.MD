# releasefeed

```text
README: Written by AI (hate doing this).
Code: Mostly written by hand! I'm using this project to learn Golang, 
      so the code is probably a bit messy.

Want to contribute? Awesome! But please follow these two rules:
1. No AI-generated PRs: Please write your code by hand.
2. Explain your changes: Clarify exactly what your code does. I need to understand it to learn!

GitHub release aggregator. Polls GitHub Atom feeds, filters the content and publishes a single combined Atom feed for RSS readers (e.g. FreshRSS). Optionally sends Apprise notifications when a new release matches alert keywords.

## Endpoints

| Endpoint | Description |
|---|---|
| `GET /feed` | Combined Atom feed with all releases |
| `GET /release/{name}/{title}` | Plain HTML page for one release (e.g. `/release/Hyprland/v0.55.4`) — feed entry links point here so reader-mode scraping gets the filtered content instead of github.com clutter; includes a "View on GitHub" link |
| `GET /status` | JSON status per feed: last successful fetch, entry count, last error, consecutive failures |
| `GET /image/{file}` | Serves images from `config/image/` |
| `GET /` | Health check — returns `ok` |

## File layout

```
releasefeed/
├── config/
│   ├── config.yaml       # configuration
│   ├── state.json        # notification state (auto-created on first run)
│   └── image/            # per-feed images (webp, png, svg)
├── internal/
├── Dockerfile
├── docker-compose.yml
└── main.go
```

## Configuration

### server

| Field | Description |
|---|---|
| `port` | HTTP port (default `880`) |
| `base_url` | Public server address — used to build image URLs |

### apprise

| Field | Description |
|---|---|
| `url` | Apprise API URL |
| `tags` | Default notification tags |

### defaults

| Field | Description |
|---|---|
| `fetch_interval` | How often to poll feeds (e.g. `30m`, `1h`, `6h`) |
| `max_entries` | Cap on releases shown per feed in `/feed` (default `10`) — each poll replaces the feed's entries with the N newest, older ones drop off |
| `alert_keywords` | Keywords that trigger a notification |
| `exclude_sections` | Sections stripped from every feed |
| `ignore_titles` | Case-insensitive regexes — matching releases are skipped entirely (e.g. `beta`, `-rc`) |

### feeds

| Field | Required | Description |
|---|---|---|
| `name` | yes | Display name used in the feed |
| `url` | yes | GitHub Atom feed URL (`/releases.atom`) |
| `fetch_interval` | no | Overrides `defaults.fetch_interval` |
| `max_entries` | no | Overrides `defaults.max_entries` |
| `image` | no | File name from `config/image/` |
| `exclude_sections` | no | Extra sections to strip — merged with `defaults.exclude_sections` |
| `regex_remove` | no | Regular expressions — matching HTML fragments are removed |
| `alert_keywords` | no | Merged with `defaults.alert_keywords` |
| `ignore_titles` | no | Merged with `defaults.ignore_titles` |
| `notify_all` | no | `true` = notify on every new release, not only on keyword match |
| `apprise_tags` | no | Overrides global `apprise.tags` |

## Merging defaults

`alert_keywords` and `exclude_sections` are always **merged** — per-feed values add to the defaults instead of replacing them. Duplicates are removed.

```yaml
defaults:
  alert_keywords:
    - "breaking"
    - "security"
  exclude_sections:
    - "sponsors"
    - "donators"

feeds:
  - name: "Hyprland"
    alert_keywords:
      - "critical"          # effective: breaking, security, critical
    exclude_sections:
      - "one-time donators" # effective: sponsors, donators, one-time donators
```

## Section filtering

Sections are detected from `<h1>`–`<h6>` tags and bold-paragraph pseudo-headings (`<p><strong>Sponsors</strong></p>`) in the HTML GitHub ships. Matching is case-insensitive and partial — `"sponsors"` also removes `"Diamond Sponsors"`.

Removing a section also removes its subsections: cutting an `<h2>` drops everything up to the next heading of the same or higher level, including any `<h3>`/`<h4>` inside.

Text before the first section (the preamble) is always kept.

## Feed stability

Editing a release on GitHub bumps its `<updated>` timestamp, which would resurface old releases as new in RSS readers. releasefeed pins the timestamp it first saw for each release, so edits upstream don't reorder the feed.

## Apprise notifications

A notification is sent when a release is new (not present in `state.json`) and its title or content contains one of the `alert_keywords`, or the feed has `notify_all: true`.

Notification content: `[FeedName] version` title plus a link to the release.

On a fresh (empty) `state.json` the first poll only records existing releases without notifying, so a redeploy never floods you with alerts about old releases.

When a feed fails to fetch 5 times in a row, a single Apprise notification is sent so a silently broken feed doesn't go unnoticed. The failure counter resets on the next successful fetch and is visible under `/status`.

## Running

```bash
docker compose up -d
docker logs releasefeed -f
```
