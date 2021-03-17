package amplicons

import (
	"testing"
)

var (
	manifestURL       = "https://raw.githubusercontent.com/artic-network/primer-schemes/master/schemes_manifest.json"
	primersURL        = "https://github.com/artic-network/primer-schemes/raw/master/nCoV-2019/V3/nCoV-2019.primer.bed"
	refURL            = "https://github.com/artic-network/primer-schemes/raw/master/nCoV-2019/V3/nCoV-2019.reference.fasta"
	numSCOV2amplicons = 98
	queryRead         = []byte("CAAACCAACCAACTTTCGATCTCTTGTAGATCTGTTCTCTAAACGAACTTTAAAATCTGTGTGGCTGTCACTCGGCTGCATGCTTAGTGCACTCACGCAGTATAATTAATAACTAATTACTGTCGTTGACAGGACACGAGTAACTCGTCTATCTTCTGCAGGCTGCTTACGGTTTCGTCCGTGTTGCAGCCGATCATCAGCACATCTAGGTTTCGTCCGGGTGTGACCGAAAGGTAAGATGGAGAGCCTTGTCCCTGGTTTCAACGAGAAAACACACGTCCAACTCAGTTTGCCTGTTTTACAGGTTCGCGACGTGCTCGTACGTGGCTTTGGAGACTCCGTGGAGGAGGTCTTATCAGAGGCACGTCAACATCTTAAAGATGGCACTTGTGGCTTAGTAGAAGTTGAAAAAGGCGTTTTGCCTCAACTTGAACAGCCCTATGTGTTCAT")
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
	if a.GetMeanSize() != 393 {
		t.Fatalf("did not get correct mean amplicon size: wanted 393, got %d", a.GetMeanSize())
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
		if len(amplicon.sketch.GetSketch()) != SketchSize {
			t.Fatalf("sketch was not expected size: wanted %d, got %d", SketchSize, len(amplicon.sketch.GetSketch()))
		}
	}
}

// TestGetTopHit
func TestGetTopHit(t *testing.T) {
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
	}
	topHit, _, err := a.GetTopHit(queryRead)
	if err != nil {
		t.Fatal(err)
	}
	if topHit != "1" {
		t.Logf("incorrect amplicon return for top hit: wanted 1, got %s", topHit)
	}
}
