# lanternfish

The lantern desk for the Harbour Steps night market.
Stallholders borrow lanterns at dusk and bring them back at close,
and lanternfish keeps track of which lantern is with whom.

It is one small HTTP service with the inventory held in memory.
Every change is written to a JSON snapshot so the inventory survives a restart.

## Running

```sh
go run . -seed data/seed.json
```

The first start loads `data/seed.json`.
After that the snapshot at `data/snapshot.json` wins.

## API

| method | path | does |
| --- | --- | --- |
| `GET` | `/lanterns` | every lantern, sorted by id |
| `GET` | `/lanterns/{id}` | one lantern |
| `POST` | `/lanterns` | add a lantern: `{"id": "...", "name": "...", "colour": "..."}` |
| `POST` | `/lanterns/{id}/checkout` | lend it out: `{"borrower": "..."}` |
| `POST` | `/lanterns/{id}/return` | bring it back |

```sh
curl -s localhost:8086/lanterns/jade-01/checkout -d '{"borrower": "Noodle stall 4"}'
```

## Development

```sh
go test ./...
```
