package minhash

import (
	"testing"
)

var (
	sketchSize = 5
)

// TestMinHash
func TestMinHash(t *testing.T) {
	mh := New(1, sketchSize)
	valChan := make(chan uint64)
	go func() {
		for i := uint64(0); i < uint64(sketchSize*2); i++ {
			valChan <- i
		}
		close(valChan)
	}()
	mh.Add(valChan)
	sketch := mh.GetSketch()
	for i := uint64(0); i < uint64(sketchSize); i++ {
		if sketch[i] != i {
			t.Fatalf("sketch does not contain expected minimums: wanted %d, got %d", i, sketch[i])
		}
	}
}
