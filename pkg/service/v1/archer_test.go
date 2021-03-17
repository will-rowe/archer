package service

import (
	"os"
	"testing"
)

var (
	dbLocation  string = "./tmp"
	manifestURL string = "https://raw.githubusercontent.com/artic-network/primer-schemes/master/schemes_manifest.json"
)

// cleanUp is called to remove the database
// after testing completes
func cleanUp() error {
	return os.RemoveAll(dbLocation)
}

// TestAPIversion will check that API version requests
// are handled appropriately.
func TestAPIversion(t *testing.T) {
	v1 := "1"
	v2 := "2"
	aInterface, shutdown, err := NewArcher(SetDb(dbLocation))
	var a *Archer = aInterface.(*Archer)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.checkAPI(v1); err != nil {
		t.Fatal(err)
	}
	if err := a.checkAPI(v2); err == nil {
		t.Fatal("unsupported API missed by service API check")
	}
	if err := shutdown(); err != nil {
		t.Fatal(err)
	}
}

// TestPrimerScheme will make sure the manifest and individual
// primer schemes can be downloaded.
func TestPrimerScheme(t *testing.T) {
	aInterface, shutdown, err := NewArcher(SetDb(dbLocation), SetManifest(manifestURL))
	var a *Archer = aInterface.(*Archer)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: check each scheme in the manifest can be download
	_ = a

	// final test so remove the tmp db
	if err := shutdown(); err != nil {
		t.Fatal(err)
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}
