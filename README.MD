# releasefeed

Agregator release'ów z GitHub. Pobiera GitHub Atom feeds, filtruje treść i publikuje jeden zbiorczy Atom feed dla czytników RSS (np. FreshRSS). Opcjonalnie wysyła powiadomienia przez Apprise przy wykryciu słów kluczowych.

## Endpoints

| Endpoint | Opis |
|---|---|
| `GET /feed` | Atom feed ze wszystkimi release'ami |
| `GET /image/{plik}` | Serwowanie obrazków z `config/images/` |
| `GET /` | Health check — zwraca `ok` |

## Struktura plików

```
releasefeed/
├── config/
│   ├── config.yaml       # konfiguracja
│   ├── state.json        # stan powiadomień (echo '{}' > config/state.json)
│   └── images/           # obrazki per feed (webp, png, svg)
├── internal/
├── Dockerfile
├── docker-compose.yml
└── main.go
```

## Konfiguracja

### server

| Pole | Opis |
|---|---|
| `port` | Port HTTP (domyślnie `8080`) |
| `base_url` | Publiczny adres serwera — używany do budowania URL obrazków |

### apprise

| Pole | Opis |
|---|---|
| `url` | URL do Apprise API |
| `tags` | Domyślne tagi dla powiadomień |

### defaults

| Pole | Opis |
|---|---|
| `fetch_interval` | Jak często pobierać feed (np. `30m`, `1h`, `6h`) |
| `alert_keywords` | Słowa kluczowe triggerujące powiadomienie |
| `exclude_sections` | Sekcje do wycięcia z każdego feeda |

### feeds

| Pole | Wymagane | Opis |
|---|---|---|
| `name` | tak | Nazwa wyświetlana w feedzie |
| `url` | tak | URL do GitHub Atom feed (`/releases.atom`) |
| `fetch_interval` | nie | Nadpisuje `defaults.fetch_interval` |
| `image` | nie | Nazwa pliku z `config/images/` |
| `exclude_sections` | nie | Dodatkowe sekcje do wycięcia — mergowane z `defaults.exclude_sections` |
| `regex_remove` | nie | Wyrażenia regularne — pasujące fragmenty HTML są usuwane |
| `alert_keywords` | nie | Mergowane z `defaults.alert_keywords` |
| `apprise_tags` | nie | Nadpisuje globalne `apprise.tags` |

## Merge defaults

`alert_keywords` i `exclude_sections` są zawsze **mergowane** — per feed dodaje do defaults, nie zastępuje. Duplikaty są usuwane.

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
      - "critical"          # razem: breaking, security, critical
    exclude_sections:
      - "one-time donators" # razem: sponsors, donators, one-time donators
```

## Filtrowanie sekcji

Sekcje wykrywane są na podstawie tagów `<h1>`–`<h6>` w HTML z GitHuba. Dopasowanie jest case-insensitive i częściowe — `"sponsors"` wycina `"Diamond Sponsors"`.

Tekst przed pierwszą sekcją (preambuła) jest zawsze zachowany.

## Powiadomienia Apprise

Powiadomienie wysyłane gdy:
- release jest nowy (nie widziany w `state.json`)
- tytuł lub treść zawiera słowo kluczowe z `alert_keywords`

Treść powiadomienia: tytuł `[NazwaFeeda] wersja` + link do release'a.

## Uruchomienie

```bash
echo '{}' > config/state.json
docker compose up -d
docker logs releasefeed -f
```