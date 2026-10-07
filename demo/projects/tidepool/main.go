package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"tidepool/internal/tide"
)

const usage = `usage: tidepool [-data DIR] <command> [args]

commands:
  stations                  list known stations
  next <station>            the next high and low water
  table <station> <date>    hourly heights for a date (YYYY-MM-DD)
`

var errUsage = errors.New("bad usage")

func main() {
	if err := run(os.Args[1:], os.Stdout, time.Now); err != nil {
		if errors.Is(err, errUsage) {
			fmt.Fprint(os.Stderr, usage)
		} else {
			fmt.Fprintln(os.Stderr, "tidepool:", err)
		}
		os.Exit(1)
	}
}

func run(args []string, out io.Writer, now func() time.Time) error {
	fs := flag.NewFlagSet("tidepool", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dataDir := fs.String("data", envOr("TIDEPOOL_DATA", "data"), "directory holding stations.csv and constituents.csv")
	if err := fs.Parse(args); err != nil {
		return errUsage
	}
	args = fs.Args()
	if len(args) == 0 {
		return errUsage
	}

	stations, err := tide.LoadDir(*dataDir)
	if err != nil {
		return err
	}

	switch args[0] {
	case "stations":
		for _, id := range stations.IDs() {
			st := stations[id]
			fmt.Fprintf(out, "%-16s %s (%s)\n", id, st.Name, st.Location)
		}
		return nil

	case "next":
		if len(args) != 2 {
			return errUsage
		}
		st, err := stations.Get(args[1])
		if err != nil {
			return err
		}
		for _, e := range st.NextExtremes(now(), 2) {
			fmt.Fprintf(out, "%-4s  %s  %5.2f m\n", e.Kind, e.Time.In(st.Location).Format("Mon 02 Jan 15:04"), e.Height)
		}
		return nil

	case "table":
		if len(args) != 3 {
			return errUsage
		}
		st, err := stations.Get(args[1])
		if err != nil {
			return err
		}
		day, err := time.Parse("2006-01-02", args[2])
		if err != nil {
			return fmt.Errorf("date %q: want YYYY-MM-DD", args[2])
		}
		fmt.Fprintf(out, "%s, %s\n", st.Name, day.Format("Monday 2 January 2006"))
		for _, r := range st.Table(day, time.Hour, 24) {
			fmt.Fprintf(out, "%s  %5.2f m\n", r.Time.In(st.Location).Format("15:04"), r.Height)
		}
		return nil
	}
	return errUsage
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
