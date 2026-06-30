package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/paulmach/osm/osmpbf"
)

func main() {
	file, err := os.Open("osm/minnesota-260629.osm.pbf")
	if err != nil {
		log.Fatalf("opening osm file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("closing file: %v", err)
		}
	}()

	scanner := osmpbf.New(context.Background(), file, 4)
	defer func() {
		if err := scanner.Close(); err != nil {
			log.Printf("closing scanner: %v", err)
		}
	}()

	header, err := scanner.Header()
	if err != nil {
		log.Fatalf("scanning header: %v", err)
	} else {
		b, _ := json.MarshalIndent(header, "", "  ")
		log.Printf("header:\n%s", b)
	}

	i := 0
	for scanner.Scan() {
		i++
		if i == 2 {
			break
		}

		obj := scanner.Object()
		b, _ := json.MarshalIndent(obj, "", "  ")
		log.Printf("object #%d\n%s", i, b)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanning: %v", err)
	}
}
