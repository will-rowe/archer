package amplicons

import (
	"testing"
)

var (
	manifestURL = "https://raw.githubusercontent.com/artic-network/primer-schemes/master/schemes_manifest.json"
)

// TestGetManifest
func TestGetManifest(t *testing.T) {
	if _, err := GetManifest(manifestURL); err != nil {
		t.Fatal(err)
	}
}

// TestCheckManifest
func TestCheckManifest(t *testing.T) {
	man, err := GetManifest(manifestURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckManifest(man, "ebola", 1); err != nil {
		t.Fatal(err)
	}
}
