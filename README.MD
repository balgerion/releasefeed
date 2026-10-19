# releasefeed

```text
README: Written by AI (hate doing this).
Code: Mostly written by hand! I'm using this project to learn Golang, 
      so the code is probably a bit messy nd definitely not as "perfect" as what an AI would spit out.

Want to contribute? Awesome! But please follow these two rules:
1. No AI-generated PRs: Please write your code by hand.
2. Explain your changes: Clarify exactly what your code does. I need to understand it to learn!
```

GitHub release aggregator. Polls GitHub Atom feeds, filters the content and publishes a single combined Atom feed for RSS readers (e.g. FreshRSS). Optionally sends Apprise notifications when a new release matches alert keywords.

## Endpoints

| Endpoint | Description |
|---|---|
| `GET /feed` | Combined Atom feed with all releases |
| `GET /release/{name}/{title}` | Minimal HTML page for one release (e.g. `/release/Hyprland/v0.55.4`) — feed entry links point here so reader-mode scraping gets the filtered content instead of github.com clutter; includes a "View on GitHub" link, follows the system light/dark theme |
| `GET /api/releases` | JSON list of the releases currently shown in `/feed`, newest first, with `feed`, `title`, `url`, `published` and the filtered `content`. `?feed=<name>` keeps one feed (case-insensitive), `?q=<text>` keeps releases whose title or content contains the text (case-insensitive); both combine, e.g. `/api/releases?feed=Renovate&q=breaking` |
| `GET /status` | JSON status per feed: last successful fetch, entry count, last error, consecutive failures |
| `GET /image/{file}` | Serves images from `config/image/` |
| `GET /` | Health check — returns `ok` |

## File layout

```
releasefeed/
├── config/
│   ├── config.yaml       # configuration
│   ├── state.json        # notification state (auto-created on first run)
│   ├── alerted.json      # releases already notified (auto-created on first run)
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

## Images

Each entry carries its image as `<media:thumbnail>`, as a small `<img>` at the top of the content, and as an Atom `<link rel="enclosure">` with the MIME type taken from the file extension. Readers that only look at enclosures for thumbnails (FreshRSS exposes enclosures through its Google Reader API, not `media:thumbnail`) pick it up from there.

GitHub ships release-note images as `private-user-images.githubusercontent.com` URLs signed with a JWT that expires after a few minutes, so they always show up broken in RSS readers. releasefeed rewrites them to the stable `github.com/user-attachments/assets/...` form, which GitHub redirects to a freshly signed URL on every request.

## Feed stability

Editing a release on GitHub bumps its `<updated>` timestamp, which would resurface old releases as new in RSS readers. releasefeed pins the timestamp it first saw for each release, so edits upstream don't reorder the feed.

## Apprise notifications

`state.json` holds, for every feed, exactly the release IDs that feed returned on its last successful poll, with the timestamp first seen for each. Its size follows the feeds themselves, so nothing is ever evicted while a feed still lists it.

A notification is sent once per release: when its title or content contains one of the `alert_keywords`, or when it is new and the feed has `notify_all: true`. The keywords are checked on every poll for every release the feed still lists, not only on the first one, because projects often publish a release first and add the "Breaking changes" section to its notes later (qui v1.33.0 on 2026-10-19). `alerted.json` next to `state.json` records, per feed, which of those releases were already notified, so an edit that keeps or adds keywords never sends a second notification. It follows the feed the same way `state.json` does.

When `alerted.json` does not exist yet, every release already in `state.json` counts as notified, so the first start of this version sends nothing about releases it had seen before.

Notification content: `[FeedName] version` title plus a link to the release.

The first poll of a feed that has no entry in `state.json` yet (a newly added feed, or a fresh state file) only records its releases without notifying, so adding feeds or redeploying never floods you with alerts about old releases. A `state.json` in the older flat format is discarded on start in the same way.

A release whose notification fails is not marked as notified, so the next poll sends it again until Apprise accepts it.

When a feed fails to fetch 5 times in a row, a single Apprise notification is sent so a silently broken feed doesn't go unnoticed. The failure counter resets on the next successful fetch and is visible under `/status`.

## Running

The HTTP server starts before the first poll, so `/` answers at once and `/feed`, `/api/releases` and `/status` fill in feed by feed while the first round runs.

Locally, with the bundled `docker-compose.yml`:

```bash
docker compose up -d
docker logs releasefeed -f
```

In the cluster it runs as the Deployment `releasefeed` in `applications/releasefeed` of `balgeriada/k3s`. `config.yaml` lives in that directory and reaches the Pod as a ConfigMap mounted over `config/config.yaml`, so a change to the feeds is a pull request there and rolls the Pod; `state.json` and `image/` stay on the claim `releasefeed-config` mounted at `config/`. Notifications go to the `releasefeed` route of `apprise-gw`, which hands them to Apprise.

## Building

The `build-docker` workflow (`.forgejo/workflows/build-docker.yml`, started by hand) builds and pushes `git.balgeriada.com/balgerion/releasefeed:latest`, then posts a cloudevent to the Argo CD Image Updater webhook with the Actions secret `IMAGE_UPDATER_WEBHOOK`, so the new digest is committed to `balgeriada/k3s` within minutes instead of with the nightly run. Without the secret the step only warns.
