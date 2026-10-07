# kiln

Reads the CSV log a kiln controller writes during a firing and says what happened.
It reports the peak temperature, how long the kiln held near it, the rate of each heating and cooling segment,
and anything that looks wrong, such as a thermocouple spike or a gap in the log.

## Usage

```sh
python -m kiln report samples/bisque-0914.csv
```

Or install it and use the `kiln` command:

```sh
pip install -e '.[test]'
kiln report samples/glaze-0921.csv
```

## Log format

One reading per row with a header line:

```csv
timestamp,temperature_c,cone
2026-09-14 06:00:00,20.4,04
```

The `cone` column is the cone the program is firing to, and may be blank.

## Samples

| file | what it is |
| --- | --- |
| `samples/bisque-0914.csv` | a cone 04 bisque, with a thermocouple glitch at 09:23 and a logger dropout around 11:10 |
| `samples/glaze-0921.csv` | a cone 6 glaze firing with a twenty-minute hold, logged every five minutes |

## Development

```sh
python -m pytest -q
```
