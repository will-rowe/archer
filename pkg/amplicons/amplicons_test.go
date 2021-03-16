package amplicons

import (
	"testing"
)

var (
	manifestURL       = "https://raw.githubusercontent.com/artic-network/primer-schemes/master/schemes_manifest.json"
	primersURL        = "https://github.com/artic-network/primer-schemes/raw/master/nCoV-2019/V3/nCoV-2019.primer.bed"
	refURL            = "https://github.com/artic-network/primer-schemes/raw/master/nCoV-2019/V3/nCoV-2019.reference.fasta"
	numSCOV2amplicons = 98
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
	tag, err := CheckManifest(man, "scov2", 1)
	if err != nil {
		t.Fatal(err)
	}
	if tag != "sars-cov-2" {
		t.Fatalf("incorrrect tag returned from manifest: wanted sars-cov-2, got %v", tag)
	}
}

// TestGetPrimers
func TestGetPrimers(t *testing.T) {
	a := make(AmpliconSet)
	if err := getPrimers(a, primersURL); err != nil {
		t.Fatal(err)
	}
	if len(a) != numSCOV2amplicons {
		t.Fatalf("incorrect number of amplicons generated from SARS-COV-2 (v3) scheme: wanted 98, got %d", len(a))
	}
}

// TestGetSequence
func TestGetSequence(t *testing.T) {
	a := make(AmpliconSet)
	if err := getPrimers(a, primersURL); err != nil {
		t.Fatal(err)
	}
	if err := getSequence(a, refURL); err != nil {
		t.Fatal(err)
	}
	for name, amp := range a {
		if len(amp.sequence) == 0 {
			t.Fatalf("no sequence produced for amplicon: %v", name)
		}
	}
}

// TestGetSketch
func TestGetSketch(t *testing.T) {
	a := make(AmpliconSet)
	if err := getPrimers(a, primersURL); err != nil {
		t.Fatal(err)
	}
	if err := getSequence(a, refURL); err != nil {
		t.Fatal(err)
	}
	for _, amplicon := range a {
		if err := amplicon.getSketch(); err != nil {
			t.Fatal(err)
		}
		if len(amplicon.sketch) != sketchSize {
			t.Fatalf("sketch was not expected size: wanted %d, got %d", sketchSize, len(amplicon.sketch))
		}
	}
}
