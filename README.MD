# releasefeed

Agregator release'ów z GitHub. Pobiera GitHub Atom feeds, filtruje treść i publikuje jeden zbiorczy Atom feed dla czytników RSS (np. FreshRSS). Opcjonalnie wysyła powiadomienia przez Apprise przy wykryciu słów kluczowych.

## Endpoints

| Endpoint | Opis |
|---|---|
| `GET /feed` | Atom feed ze wszystkimi release'ami |
| `GET /image/{plik}` | Serwowanie obrazków z `config/images/` |
| `GET /` | Zwraca `ok` — health check |

## Struktura plików

```
releasefeed/
├── config/
│   ├── config.yaml       # konfiguracja
│   ├── state.json        # stan powiadomień (tworzyć ręcznie: echo '{}' > state.json)
│   └── images/           # obrazki per feed (webp, png, svg)
├── internal/
├── Dockerfile
├── docker-compose.yml
└── main.go
```

## Konfiguracja

### server

```yaml
server:
  port: "8080"
  base_url: "https://feed.example.com"
```

| Pole | Opis |
|---|---|
| `port` | Port HTTP (domyślnie `8080`) |
| `base_url` | Publiczny adres serwera — używany do budowania URL obrazków |

### apprise

```yaml
apprise:
  url: "https://apprise.example.com/cfg/TOKEN"
  tags:
    - releases
```

| Pole | Opis |
|---|---|
| `url` | URL do Apprise API |
| `tags` | Domyślne tagi dla powiadomień |

### defaults

```yaml
defaults:
  fetch_interval: "1h"
  alert_keywords:
    - "breaking"
    - "security"
```

| Pole | Opis |
|---|---|
| `fetch_interval` | Jak często pobierać feed (np. `30m`, `1h`, `6h`) |
| `alert_keywords` | Słowa kluczowe triggerujące powiadomienie — stosowane gdy feed nie ma własnych |

### feeds

```yaml
feeds:
  - name: "Hyprland"
    url: "https://github.com/hyprwm/Hyprland/releases.atom"
    fetch_interval: "30m"
    image: "hyprland.svg"
    sections:
      mode: exclude
      names:
        - "special thanks"
        - "sponsors"
    regex_remove:
      - "(?i)<p>made with.*?</p>"
    alert_keywords:
      - "breaking"
    apprise_tags:
      - releases
```

| Pole | Wymagane | Opis |
|---|---|---|
| `name` | tak | Nazwa wyświetlana w feedzie |
| `url` | tak | URL do GitHub Atom feed (`.../releases.atom`) |
| `fetch_interval` | nie | Nadpisuje `defaults.fetch_interval` |
| `image` | nie | Nazwa pliku z `config/images/` wyświetlana jako miniatura |
| `sections.mode` | nie | `include` — pokaż tylko wymienione sekcje / `exclude` — ukryj wymienione sekcje |
| `sections.names` | nie | Lista nazw sekcji (dopasowanie case-insensitive, częściowe) |
| `regex_remove` | nie | Lista wyrażeń regularnych — pasujące fragmenty HTML są usuwane |
| `alert_keywords` | nie | Nadpisuje `defaults.alert_keywords` dla tego feeda |
| `apprise_tags` | nie | Nadpisuje globalne `apprise.tags` dla tego feeda |

## Filtrowanie sekcji

Sekcje są wykrywane na podstawie tagów `<h1>`–`<h6>` w HTML release notes z GitHuba.

**exclude** — usuwa sekcje których nagłówek zawiera jedną z podanych nazw:
```yaml
sections:
  mode: exclude
  names:
    - "special thanks"
    - "sponsors"
    - "donators"
```

**include** — zostawia tylko sekcje których nagłówek zawiera jedną z podanych nazw:
```yaml
sections:
  mode: include
  names:
    - "what's new"
    - "bug fixes"
    - "breaking changes"
```

Tekst przed pierwszą sekcją (preambuła) jest zawsze zachowany.

## Powiadomienia Apprise

Powiadomienie jest wysyłane gdy:
- release jest nowy (nie widziany wcześniej w `state.json`)
- tytuł lub treść zawiera jedno ze słów kluczowych z `alert_keywords`

Powiadomienie zawiera tytuł w formacie `[NazwaFeeda] wersja` oraz link do release'a.

## Uruchomienie

```bash
echo '{}' > config/state.json

docker compose up -d
```

Logi:
```bash
docker logs releasefeed -f
```

## docker-compose.yml

```yaml
services:
  releasefeed:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
```