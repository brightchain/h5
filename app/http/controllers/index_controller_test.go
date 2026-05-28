package controllers

import (
	"strings"
	"testing"
)

func TestReadUniqueCitiesSplitsChineseCityNames(t *testing.T) {
	input := "开通城市\n张家港市汕头市汕头市北京市，黔南布依族苗族自治州西安市"

	cities, err := readUniqueCities(strings.NewReader(input))
	if err != nil {
		t.Fatalf("readUniqueCities returned error: %v", err)
	}

	want := []string{"张家港市", "汕头市", "北京市", "黔南布依族苗族自治州", "西安市"}
	if len(cities) != len(want) {
		t.Fatalf("got %d cities %v, want %d %v", len(cities), cities, len(want), want)
	}
	for i := range want {
		if cities[i] != want[i] {
			t.Fatalf("cities[%d] = %q, want %q; all cities: %v", i, cities[i], want[i], cities)
		}
	}
}
