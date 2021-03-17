// Package amplicons performs the amplicon scheme validation, handling and filtering for Archer.
package amplicons

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/grailbio/bio/encoding/fasta"
	log "github.com/sirupsen/logrus"
	"github.com/will-rowe/nthash"

	api "github.com/will-rowe/archer/pkg/api/v1"
	"github.com/will-rowe/archer/pkg/minhash"
)

const (
	KmerSize   = 7
	SketchSize = 24
	Canonical  = true
)

// Amplicon stores the minimal information needed
// by Archer to perform FASTQ filtering for
// amplicon enrichment.
type Amplicon struct {
	refName  string           // reference sequence ID
	start    int              // start of leftmost primer (0-based indexing)
	end      int              // end of rightmost primer
	sequence []byte           // reference sequence for amplicon (incl. primer sequences)
	sketch   *minhash.MinHash // minhash sketch for the amplicon
}

// AmpliconSet
type AmpliconSet map[string]*Amplicon

// GetMeanSize returns the mean amplicon size.
// Primers and inserts are included.
func (as AmpliconSet) GetMeanSize() int {
	meanSize := 0
	for _, amplicon := range as {
		meanSize += (amplicon.end - amplicon.start)
	}
	meanSize /= len(as)
	return meanSize
}

// GetTopHit will compare a read against a each amplicon
// in the set and return the name of the amplicon with
// the best match, plus the Jaccard distance and any
// error.
func (as AmpliconSet) GetTopHit(read []byte) (string, float64, error) {

	// sketch the query read
	sketcher := minhash.New(KmerSize, SketchSize)
	hasher, err := nthash.NewHasher(&read, KmerSize)
	if err != nil {
		return "", 0.0, err
	}
	sketcher.Add(hasher.Hash(Canonical))

	// compare the query against each amplicon
	// TODO: this is inefficient and we could/should use an index
	topDist := 0.0
	topHit := ""
	for ampliconName, amplicon := range as {
		dist, err := sketcher.GetDistance(amplicon.sketch)
		if err != nil {
			return "", 0.0, err
		}
		if dist > topDist {
			topDist = dist
			topHit = ampliconName
		}
	}
	return topHit, topDist, nil
}

// NewAmpliconSet downloads the primer set and
// reference sequence for a primer scheme in
// the manifest and returns an AmpliconSet.
func NewAmpliconSet(manifest *api.Manifest, requestedScheme string, requestedVersion int32) (*AmpliconSet, error) {
	a := make(AmpliconSet)

	// download primers and create amplicons
	primersURL := manifest.Schemes[requestedScheme].PrimerUrls[strconv.Itoa(int(requestedVersion))]
	log.Tracef("downloading and parsing primers from: %v", primersURL)
	if err := getPrimers(a, primersURL); err != nil {
		return nil, err
	}

	// download reference sequence and add seqs to amplicons
	refURL := manifest.Schemes[requestedScheme].ReferenceUrls[strconv.Itoa(int(requestedVersion))]
	log.Tracef("downloading reference sequence and extracting amplicons from: %v", refURL)
	if err := getSequence(a, refURL); err != nil {
		return nil, err
	}

	// create amplicon sketches
	log.Tracef("creating sketches for %d amplicons", len(a))
	var wg sync.WaitGroup
	wg.Add(len(a))
	for _, amplicon := range a {
		go func(wg *sync.WaitGroup, amplicon *Amplicon) {
			defer wg.Done()
			amplicon.getSketch()
		}(&wg, amplicon)
	}
	wg.Wait()

	return &a, nil
}

// getSketch will generate a minhash sketch for an amplicon.
func (a *Amplicon) getSketch() error {

	// create a minhash sketcher
	a.sketch = minhash.New(KmerSize, SketchSize)

	// create the ntHash iterator using a pointer to the sequence and a k-mer size
	hasher, err := nthash.NewHasher(&a.sequence, KmerSize)

	// check for errors (e.g. bad k-mer size choice)
	if err != nil {
		return err
	}

	// attach the hasher to the sketcher and populate the sketch
	a.sketch.Add(hasher.Hash(Canonical))

	return nil
}

// getPrimers collects primers from a URL and constructs amplicons.
// It populates the provided map and returns any error.
func getPrimers(as AmpliconSet, url string) error {

	// download the primer set
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// set up a tsv reader
	reader := csv.NewReader(resp.Body)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	// create amplicons from primers using ARTIC pipeline logic
	for {

		// get the primer line and check errs/end of file
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// get amplicon name
		ampliconName := strings.Split(line[3], "_")[1]

		// update or create amplicon in set
		amplicon, ok := as[ampliconName]
		if !ok {
			amplicon = &Amplicon{
				refName:  line[0],
				start:    99999999,
				end:      0,
				sequence: nil,
				sketch:   nil,
			}
			as[ampliconName] = amplicon
		}

		// detect primer orientation and update amplicon boundaries (accounts for alts)
		switch line[5] {
		case "+":
			start, err := strconv.Atoi(line[1])
			if err != nil {
				return err
			}
			if start < amplicon.start {
				amplicon.start = start
			}
		case "-":
			end, err := strconv.Atoi(line[2])
			if err != nil {
				return err
			}
			if end > amplicon.end {
				amplicon.end = end
			}
		default:
			return fmt.Errorf("can't detect primer orientation from in %v", line)
		}
	}
	return nil
}

// getSequence collects reference sequence from a URL
// updates amplicons in the provided map to include
// their sequence. It returns any errror.
// It populates the provided map and returns any error.
func getSequence(as AmpliconSet, url string) error {

	// download the reference sequence file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// split into two readers
	var fastaBuf bytes.Buffer
	tee := io.TeeReader(resp.Body, &fastaBuf)

	// index the fasta
	idx := bytes.Buffer{}
	if err := fasta.GenerateIndex(&idx, tee); err != nil {
		return err
	}

	// read fasta
	fa, err := fasta.NewIndexed(bytes.NewReader(fastaBuf.Bytes()), bytes.NewReader(idx.Bytes()))
	if err != nil {
		return err
	}

	// loop over the amplicons and populate the sequence fields
	for _, amplicon := range as {
		seq, err := fa.Get(amplicon.refName, uint64(amplicon.start), uint64(amplicon.end))
		if err != nil {
			return err
		}
		amplicon.sequence = []byte(seq)
	}
	return nil
}
