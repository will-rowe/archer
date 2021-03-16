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
	kmerSize   = 7
	sketchSize = 24
	canonical  = true
)

// Amplicon stores the minimal information needed
// by Archer to perform FASTQ filtering for
// amplicon enrichment.
type Amplicon struct {
	refName  string   // reference sequence ID
	start    int      // start of leftmost primer (0-based indexing)
	end      int      // end of rightmost primer
	sequence []byte   // reference sequence for amplicon (incl. primer sequences)
	sketch   []uint64 // minhash sketch for the amplicon
}

// AmpliconSet
type AmpliconSet map[string]*Amplicon

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
	sketcher := minhash.New(kmerSize, sketchSize)

	// create the ntHash iterator using a pointer to the sequence and a k-mer size
	hasher, err := nthash.NewHasher(&a.sequence, kmerSize)

	// check for errors (e.g. bad k-mer size choice)
	if err != nil {
		return err
	}

	// attach the hasher to the sketcher and populate the sketch
	sketcher.Add(hasher.Hash(canonical))

	// collect the sketch and add it to the amplicon
	a.sketch = sketcher.GetSketch()
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
