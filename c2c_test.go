package coord2country

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex_GeoJSON_Query(t *testing.T) {
	f, err := os.Open("countries.geojsonseq")
	require.NoError(t, err)

	idx, err := OpenIndexFromGeoJSONSeq(f)
	require.NoError(t, err)

	tests := []struct {
		name string
		lat  float64
		lng  float64
		want string
	}{
		{"Paris", 48.8, 2.2, "France"},
		{"Toronto", 51.213890, -102.462776, "Canada"},
		{"Beijing", 39.916668, 116.383331, "China"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idx.Query(tt.lat, tt.lng)
			require.Len(t, got, 1)
			require.Equal(t, tt.want, got[0].Name)
		})
	}
}

func TestIndex_FromData_Query(t *testing.T) {
	f, err := os.Open("countries.dat")
	require.NoError(t, err)

	idx, err := OpenIndexFromData(f)
	require.NoError(t, err)

	tests := []struct {
		name string
		lat  float64
		lng  float64
		want string
	}{
		{"Paris", 48.8, 2.2, "France"},
		{"Toronto", 51.213890, -102.462776, "Canada"},
		{"Beijing", 39.916668, 116.383331, "China"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idx.Query(tt.lat, tt.lng)
			require.Len(t, got, 1)
			require.Equal(t, tt.want, got[0].Name)
		})
	}
}

func BenchmarkTestGeoJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open("countries.geojsonseq")
		require.NoError(b, err)

		_, err = OpenIndexFromGeoJSONSeq(f)
		require.NoError(b, err)
	}
}

func BenchmarkTestData(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open("countries.dat")
		require.NoError(b, err)

		_, err = OpenIndexFromData(f)
		require.NoError(b, err)
	}
}

func BenchmarkQuery(b *testing.B) {
	f, err := os.Open("countries.dat")
	require.NoError(b, err)

	idx, err := OpenIndexFromData(f)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		r := idx.Query(48.8, 2.2)
		if len(r) < 1 {
			b.Fail()
		}
	}
}

func BenchmarkQueryDeadZone(b *testing.B) {
	f, err := os.Open("countries.dat")
	require.NoError(b, err)

	idx, err := OpenIndexFromData(f)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		r := idx.Query(0, 0)
		if len(r) != 0 {
			b.Fail()
		}
	}
}
