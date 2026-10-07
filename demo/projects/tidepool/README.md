# tidepool

A small command-line tide predictor for the three stations along the Saltmarsh coast.
It sums a handful of harmonic constituents per station and prints the next high and low water,
or an hourly table for a day.

It ignores nodal corrections and the astronomical arguments a real predictor would use,
so treat its numbers as good to a few tens of centimetres.

## Usage

```sh
go build .
./tidepool stations
./tidepool next kelp-hollow
./tidepool table brackwater 2026-01-15
```

Data is read from `./data` by default.
Point `-data` or `TIDEPOOL_DATA` somewhere else to use another set of stations.

## Data

`data/stations.csv` lists each station with its IANA timezone and its mean water level above chart datum.
`data/constituents.csv` gives each station's constituents.

## Development

```sh
go test ./...
```
