// Package minhash is a simple KMV MinHash implementation.
package minhash

import (
	"container/heap"
	"sort"
)

// MinHash is used to process hashed k-mers
// and return a sketch of minimums.
type MinHash struct {
	kSize      int
	sketchSize int
	sketch     *MinHashSketch
}

// MinHashSketch is a max-heap of uint64s.
type MinHashSketch []uint64

func (mh MinHashSketch) Len() int           { return len(mh) }
func (mh MinHashSketch) Less(i, j int) bool { return mh[i] > mh[j] }
func (mh MinHashSketch) Swap(i, j int)      { mh[i], mh[j] = mh[j], mh[i] }
func (mh *MinHashSketch) Push(x interface{}) {
	*mh = append(*mh, x.(uint64))
}
func (mh *MinHashSketch) Pop() interface{} {
	old := *mh
	n := len(old)
	x := old[n-1]
	*mh = old[0 : n-1]
	return x
}

// New returns an initialised minhash object which
// is ready to receive hashed k-mers.
func New(kSize, sketchSize int) *MinHash {

	// create the minhash object
	mh := &MinHash{
		kSize:      kSize,
		sketchSize: sketchSize,
		sketch:     new(MinHashSketch),
	}

	// init the heap, which we use as the sketch
	heap.Init(mh.sketch)
	return mh
}

// Add will check k-mers and add minimums
// to the sketch.
func (mh *MinHash) Add(kmerChan <-chan uint64) {
	for kmer := range kmerChan {

		// if sketch isn't full, add the kmer and continue
		if len(*mh.sketch) < mh.sketchSize {
			heap.Push(mh.sketch, kmer)
			continue
		}

		// if sketch is full, check if current max needs popping
		if (*mh.sketch)[0] > kmer {

			// replace largest sketch value with a new min
			_ = heap.Pop(mh.sketch)
			heap.Push(mh.sketch, kmer)
		}
	}
}

// GetSketch returns the current sketch from
// the minhash object.
func (mh *MinHash) GetSketch() []uint64 {
	sketch := make([]uint64, len(*mh.sketch))
	for i, val := range *mh.sketch {
		sketch[i] = val
	}
	sort.Slice(sketch, func(i, j int) bool { return sketch[i] < sketch[j] })
	return sketch
}
