package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":8086", "address to listen on")
	snapshot := flag.String("snapshot", "data/snapshot.json", "JSON file the inventory is saved to and loaded from")
	seed := flag.String("seed", "", "JSON file of lanterns to load when there is no snapshot yet")
	flag.Parse()

	store := NewStore(time.Now)
	if err := store.Load(*snapshot); err != nil {
		log.Fatal(err)
	}
	if *seed != "" && len(store.List()) == 0 {
		if err := store.Load(*seed); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("lanternfish: %d lanterns, listening on %s", len(store.List()), *addr)
	log.Fatal(http.ListenAndServe(*addr, NewServer(store, *snapshot).Routes()))
}
