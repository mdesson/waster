# Waster

Waste your time less wastefully. Simple little stumebleupon-like. Click a button and it brings you to a random site. Includes adding/removing sites. 

## Installation

Keep your files in a `sites.json` with the follow format:

```json
[
  { "title": "Random Wikipedia Article", "url": "https://en.wikipedia.org/wiki/Special:Random" }
]
```

By default it is expected to be in the same directory as the binary runs, however you can override it with the env var `SITES_DIR`. Do not include a trailing slash. Example:
```env
SITES_DIR=/config
```