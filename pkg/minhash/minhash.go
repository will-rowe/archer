// Package minhash is a simple KMV MinHash implementation.
package minhash

import (
	"container/heap"
	"fmt"
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
			(*mh.sketch)[0] = kmer
			heap.Fix(mh.sketch, 0)
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

// GetDistance returns the jaccard distance
// between two sketches.
func (mh *MinHash) GetDistance(query *MinHash) (float64, error) {
	if mh.kSize != query.kSize {
		return 0.0, fmt.Errorf("kmer sizes do not match: got %d and %d", mh.kSize, query.kSize)
	}
	if mh.sketchSize != query.sketchSize {
		return 0.0, fmt.Errorf("sketch sizes do not match: got %d and %d", mh.sketchSize, query.sketchSize)
	}
	minimums := make(map[uint64]float64, len(*mh.sketch))
	for _, val := range *mh.sketch {
		minimums[val]++
	}
	intersect := 0.0
	for _, val := range *query.sketch {
		if count, ok := minimums[val]; ok && count > 0 {
			intersect++
			minimums[val] = count - 1
		}
	}
	maxLen := len(*mh.sketch)
	if maxLen < len(*query.sketch) {
		maxLen = len(*query.sketch)
	}
	return intersect / float64(maxLen), nil
}
