# Bip Card API

An api that allows get information from any bip card

Internally this api makes web-scraping to get data related an any bipcard from webpage: http://pocae.tstgo.cl/PortalCAE-WAR-MODULE/

```bash
# Run server
make run-server
```

```bash
## Examples

# Get bip card info
curl "localhost:8080/api/v1/bipcard/<CARD_NUMBER>/info"

# Get bip card movements
curl "localhost:8080/api/v1/bipcard/<CARD_NUMBER>/movements"
```
