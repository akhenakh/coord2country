package coord2country

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/golang/geo/s2"
	"github.com/peterstace/simplefeatures/geom"
)

//go:embed countries.dat
var data []byte

type Index struct {
	*s2.ShapeIndex
	*s2.ContainsPointQuery
}

type IndexedLoop struct {
	*s2.Loop
	Name, Country string
}

func NewIndex() *Index {
	shapeIndex := s2.NewShapeIndex()

	return &Index{
		ShapeIndex: shapeIndex,
	}
}

// OpenIndexFromData reads the index from a file as follow:
// uint8 <- size of string Data separated by |, DATA
// uint32 <- size of s2 polygon encoded, DATA
func OpenIndexFromData(r io.Reader) (*Index, error) {
	idx := NewIndex()

	il := IndexedLoop{}

	if err := func() error {
		for {
			var size uint16
			if err := binary.Read(r, binary.BigEndian, &size); err != nil {
				return err
			}

			stb := make([]byte, size)

			n, err := io.ReadFull(r, stb)
			if err != nil {
				return err
			}

			if n != int(size) {
				return fmt.Errorf("can't read the right amount of data")
			}

			svals := strings.Split(string(stb[:]), "|")

			var lsize uint32
			if err := binary.Read(r, binary.BigEndian, &lsize); err != nil {
				return err
			}

			lb := make([]byte, lsize)
			if _, err = io.ReadFull(r, lb); err != nil {
				return err
			}

			lr := bytes.NewReader(lb)

			l := s2.EmptyLoop()
			if err := l.Decode(lr); err != nil {
				return err
			}

			il.Name = svals[0]
			il.Country = svals[1]
			il.Loop = l
			idx.Add(il)

		}
	}(); err != io.EOF {
		return nil, err
	}

	idx.ContainsPointQuery = s2.NewContainsPointQuery(idx.ShapeIndex, s2.VertexModelOpen)

	return idx, nil
}

func OpenIndex() (*Index, error) {
	r := bytes.NewReader(data)

	return OpenIndexFromData(r)
}

func OpenIndexFromGeoJSONSeq(r io.Reader) (*Index, error) {
	idx := NewIndex()

	buf := []byte{}
	s := bufio.NewScanner(r)
	s.Buffer(buf, 4096*1024)

	var f geom.GeoJSONFeature

	for s.Scan() {
		if err := json.Unmarshal(s.Bytes(), &f); err != nil {
			return nil, err
		}
		idx.IndexFeature(f)

	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	idx.ContainsPointQuery = s2.NewContainsPointQuery(idx.ShapeIndex, s2.VertexModelOpen)

	return idx, nil
}

func OpenIndexFromGeoJSON(r io.Reader) (*Index, error) {
	var fc geom.GeoJSONFeatureCollection
	if err := json.NewDecoder(r).Decode(&fc); err != nil {
		return nil, err
	}

	idx := NewIndex()

	for _, f := range fc {
		idx.IndexFeature(f)
	}

	idx.ContainsPointQuery = s2.NewContainsPointQuery(idx.ShapeIndex, s2.VertexModelOpen)

	return idx, nil
}

func (idx *Index) Query(lat, lng float64) []IndexedLoop {
	p := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))

	shapes := idx.ContainsPointQuery.ContainingShapes(p)

	resp := make([]IndexedLoop, len(shapes))

	for i, il := range shapes {
		resp[i] = il.(IndexedLoop)
	}

	return resp
}

func (idx *Index) IndexFeature(f geom.GeoJSONFeature) {
	switch f.Geometry.Type() {
	case geom.TypePolygon:
		p := f.Geometry.MustAsPolygon().ForceCCW()
		idx.Add(IndexedLoop{
			Name:    f.Properties["ADMIN"].(string),
			Country: f.Properties["ISO_A3"].(string),
			Loop:    LoopFromPolygon(p),
		})

	case geom.TypeMultiPolygon:
		for i := 0; i < f.Geometry.MustAsMultiPolygon().NumPolygons(); i++ {
			p := f.Geometry.MustAsMultiPolygon().PolygonN(i).ForceCCW()
			idx.Add(IndexedLoop{
				Name:    f.Properties["ADMIN"].(string),
				Country: f.Properties["ISO_A3"].(string),
				Loop:    LoopFromPolygon(p),
			})
		}
	}
}

// LoopFromPolygon creates an s2 loop from the external ring (no holes)
func LoopFromPolygon(g geom.Polygon) *s2.Loop {
	seq := g.Coordinates()[0]
	c := make([]float64, seq.Length()*2)

	for i := 0; i < seq.Length(); i++ {
		xy := seq.GetXY(i)
		c[i*2] = xy.X
		c[i*2+1] = xy.Y
	}

	if len(c)%2 != 0 || len(c) < 2*3 {
		return nil
	}

	points := make([]s2.Point, len(c)/2)

	for i := 0; i < len(c); i += 2 {
		points[i/2] = s2.PointFromLatLng(s2.LatLngFromDegrees(c[i+1], c[i]))
	}

	if points[0] == points[len(points)-1] {
		points = append(points[:len(points)-1], points[len(points)-1+1:]...)
	}

	loop := s2.LoopFromPoints(points)

	return loop
}

func GeomFromLoop(l *s2.Loop) geom.Geometry {
	coords := make([]float64, (l.NumVertices()+1)*2)

	for i, p := range l.Vertices() {
		coords[i*2] = s2.LatLngFromPoint(p).Lng.Degrees()
		coords[i*2+1] = s2.LatLngFromPoint(p).Lat.Degrees()
	}

	coords[l.NumVertices()*2] = coords[0]
	coords[l.NumVertices()*2+1] = coords[1]

	seq := geom.NewSequence(coords, geom.DimXY)
	ls, _ := geom.NewLineString(seq)
	p, _ := geom.NewPolygon([]geom.LineString{ls})

	return p.AsGeometry()
}
