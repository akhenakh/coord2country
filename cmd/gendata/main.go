package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/akhenakh/coord2country"
	"github.com/golang/geo/s2"
	"github.com/peterstace/simplefeatures/geom"
)

func main() {
	geojsonPath := flag.String("geojsonPath", "countries.geojsonseq", "GeoJSON seq files with countries")
	outputPath := flag.String("ouput", "countries.dat", "Output file")

	flag.Parse()

	fo, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	fg, err := os.Open(*geojsonPath)
	if err != nil {
		panic(err)
	}
	defer fg.Close()

	buf := []byte{}
	s := bufio.NewScanner(fg)
	s.Buffer(buf, 4096*1024)

	var f geom.GeoJSONFeature

	for s.Scan() {
		if err := json.Unmarshal(s.Bytes(), &f); err != nil {
			panic(err)
		}

		switch f.Geometry.Type() {
		case geom.TypePolygon:
			p := f.Geometry.MustAsPolygon().ForceCCW()
			l := coord2country.LoopFromPolygon(p)
			err := saveLoop(l, fo, f.Properties["ADMIN"].(string), f.Properties["ISO_A2"].(string))
			if err != nil {
				panic(err)
			}
		case geom.TypeMultiPolygon:
			for i := 0; i < f.Geometry.MustAsMultiPolygon().NumPolygons(); i++ {
				p := f.Geometry.MustAsMultiPolygon().PolygonN(i).ForceCCW()
				l := coord2country.LoopFromPolygon(p)
				err := saveLoop(l, fo, f.Properties["ADMIN"].(string), f.Properties["ISO_A2"].(string))
				if err != nil {
					panic(err)
				}
			}

		}
	}
}

func saveLoop(l *s2.Loop, w io.Writer, name, iso string) error {
	var lb bytes.Buffer
	lw := bufio.NewWriter(&lb)
	if err := l.Encode(lw); err != nil {
		return err
	}

	// uint16 <- size of string Data separated by |, DATA
	// uint32 <- size of s2 polygon encoded, DATA
	buf := make([]byte, 2+len(name)+1+len(iso))

	binary.BigEndian.PutUint16(buf, uint16(len(buf)-2))
	copy(buf[2:], name)
	buf[2+len(name)] = '|'
	copy(buf[2+len(name)+1:], iso)

	if _, err := w.Write(buf); err != nil {
		return err
	}

	if err := lw.Flush(); err != nil {
		return err
	}

	ls := uint32(lb.Len())
	if err := binary.Write(w, binary.BigEndian, ls); err != nil {
		return err
	}

	_, err := lb.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}
