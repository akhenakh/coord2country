package main

import (
	"flag"
	"fmt"

	"github.com/akhenakh/coord2country"
)

func main() {
	lat := flag.Float64("lat", 0.8, "Latitude")
	lng := flag.Float64("lng", 0.2, "Longitude")

	idx, err := coord2country.OpenIndex()
	if err != nil {
		panic(err)
	}
	for {
		fmt.Printf("%#v", idx.Query(*lat, *lng))
	}
}
