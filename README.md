# coord2country

A simple solution to query for point in polygon, very fast, mainly for countries polygons but could serve any enclosed regions.

Embedding data from [Natural Earth](https://www.naturalearthdata.com/downloads/10m-cultural-vectors/), load precomputed data for fast loading with index into memory using [s2](https://s2geometry.io/).


## Usage

```go
idx, err := coord2country.OpenIndex()
if err != nil {
	panic(err)
}

fmt.Printf("%v", idx.Query(48.8, 2.2))
```

## Data

Natural Earth Countries 10M is embedded in the library.

You can use your own data, use `cmd/gendata` to create your data file.


## Speed

Around 320 ns per query, when contained in a Polygon, around 190 ns in a dead zone.

RSS Memory is around 230MB for the 10M world countries.

## Projects Using coord2country

- [geo-benthos](https://github.com/akhenakh/geo-benthos) a GIS plugin for [Benthos](https://www.benthos.dev/), a stream processing tool, to enrich stream from coordinates.
- [ovr](https://github.com/akhenakh/ovr) the optional `-tags geo` build with the `country` command returns the country of centroid.