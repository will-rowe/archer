package minhash

import (
	"testing"
)

var (
	kmerSize   = 1
	sketchSize = 6
)

// addValues will add X sequential values to the provided minhash object.
func addValues(minhash *MinHash, x int) {
	valChan := make(chan uint64)
	go func() {
		for i := uint64(0); i < uint64(x); i++ {
			valChan <- i
		}
		close(valChan)
	}()
	minhash.Add(valChan)
	return
}

// TestMinHash
func TestMinHash(t *testing.T) {
	mh := New(kmerSize, sketchSize)
	addValues(mh, sketchSize*2)
	sketch := mh.GetSketch()
	for i := uint64(0); i < uint64(sketchSize); i++ {
		if sketch[i] != i {
			t.Fatalf("sketch does not contain expected minimums: wanted %d, got %d", i, sketch[i])
		}
	}
}

// TestSimilarity
func TestSimilarity(t *testing.T) {
	mh1 := New(kmerSize, sketchSize)
	addValues(mh1, sketchSize*2)

	// test kmer size and sketch size catch
	mh2 := New(kmerSize+1, sketchSize)
	mh3 := New(kmerSize, sketchSize+1)
	if _, err := mh1.GetDistance(mh2); err == nil {
		t.Fatal("missed k-mer size mismatch")
	}
	if _, err := mh1.GetDistance(mh3); err == nil {
		t.Fatal("missed sketch size mismatch")
	}

	// test identical sketches
	mh4 := New(kmerSize, sketchSize)
	addValues(mh4, sketchSize*2)
	dist, err := mh1.GetDistance(mh4)
	if err != nil {
		t.Log(err)
	}
	if dist != 1.0 {
		t.Fatalf("incorrect distance: expected 1.0, got %f", dist)
	}

	// test non-identical sketches
	mh5 := New(kmerSize, sketchSize)
	addValues(mh5, sketchSize/2)
	dist, err = mh1.GetDistance(mh5)
	if err != nil {
		t.Log(err)
	}
	if dist != 0.5 {
		t.Fatalf("incorrect distance: expected 0.5, got %f", dist)
	}

}
