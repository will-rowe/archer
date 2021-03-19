// Package service implements the Archer service API.
package service

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/prologic/bitcask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/will-rowe/archer/pkg/amplicons"
	api "github.com/will-rowe/archer/pkg/api/v1"
	"github.com/will-rowe/archer/pkg/bucket"
)

// useSync will run sync on every bit cask transaction, improving stability at the expense of time
const useSync = true

// apiVersion sets the API version to use
const apiVersion = "1"

// lengthThreshold is the percentage above/below the mean amplicon size to set max/min length read filtering to
const lengthThreshold = 0.2

// jaccardThreshold is the minimum jaccard distance to match a read to an amplicon
const jaccardThreshold = 0.7

// ArcherOption is a wrapper struct used to pass functional
// options to the Archer constructor.
type ArcherOption func(archer *Archer) error

// Archer is an implementation of the v1.ArcherServer.
//
// It includes some extra data
// and methods to provide extra
// functionality.
type Archer struct {

	// service lock
	sync.RWMutex

	// version of API implemented by the server
	version string

	// numWorkers sets the number of process request workers to use
	numWorkers int

	// db is a key-value store for recording sample info
	db *bitcask.Bitcask

	// bucket contains the S3 info for managing uploads
	bucket *bucket.Bucket

	// manifest is the ARTIC primer scheme manifest
	manifest *api.Manifest

	// ampliconCache is an in-memory cache of the schemes used for the current session
	ampliconCache map[string]*amplicons.AmpliconSet

	// processChan is used to send incoming process requests to the process workers
	processChan chan *api.SampleInfo

	// watcherChan sends updates to any connected watcher
	watcherChan chan *api.SampleInfo
}

// SetNumWorkers is an option setter for the NewArcher
// constructor that sets the number of concurrent
// process request workers to use.
func SetNumWorkers(numWorkers int) ArcherOption {
	return func(x *Archer) error {
		if (numWorkers < 1) || (numWorkers > 12) {
			return errors.New("number of process request workers must be between 1 and 12")
		}
		x.numWorkers = numWorkers
		return nil
	}
}

// SetDb is an option setter for the NewArcher constructor
// that opens a db at the specified path and sets the
// appropriate field of the Archer struct.
func SetDb(dbPath string) ArcherOption {
	return func(x *Archer) error {

		// open/create the db and attach it
		db, err := bitcask.Open(dbPath, bitcask.WithSync(useSync))
		if err != nil {
			return err
		}
		x.db = db
		return nil
	}
}

// SetBucket is an option setter for the NewArcher constructor
// that sets the S3 bucket field of the Archer struct.
func SetBucket(name, region string) ArcherOption {
	return func(x *Archer) error {

		// create the bucket holder, check it and attach it
		b, err := bucket.New(bucket.SetName(name), bucket.SetRegion(region))
		if err != nil {
			return err
		}
		if err := b.Check(); err != nil {
			return err
		}
		x.bucket = b
		return nil
	}
}

// SetManifest is an option setter for the NewArcher constructor
// that downloads and opens a manifest and sets the appropriate
// field of the Archer struct.
func SetManifest(manifestURL string) ArcherOption {
	return func(x *Archer) error {

		// download the manifest, unpack and attach it
		manifest, err := amplicons.GetManifest(manifestURL)
		if err != nil {
			return err
		}
		x.manifest = manifest
		return nil
	}
}

// NewArcher creates the Archer server and returns
// it along with the shutdown method and any
// constructor error.
func NewArcher(options ...ArcherOption) (api.ArcherServer, func() error, error) {

	// create the service
	a := &Archer{
		version:       apiVersion,
		numWorkers:    2,
		ampliconCache: make(map[string]*amplicons.AmpliconSet),
		processChan:   make(chan *api.SampleInfo),
		watcherChan:   nil,
	}

	// set options
	for _, option := range options {
		if err := option(a); err != nil {
			return nil, nil, err
		}
	}

	// TODO: check a db is functioning
	if a.db == nil {
		return nil, nil, errors.New("dbPath is required")
	}

	// start up the process request workers
	for i := 0; i < a.numWorkers; i++ {
		go a.processWorker()
	}

	// return the instance and it's shutdown method
	return a, a.shutdown, nil
}

// checkAPI checks if requested API version is supported
// by the server.
func (a *Archer) checkAPI(requestedAPI string) error {
	if a.version != requestedAPI {
		return status.Errorf(codes.Unimplemented,
			"unsupported API version requested: current Archer service implements version '%s', but version '%s' was requested", a.version, requestedAPI)
	}
	return nil
}

// shutdown will stop the Archer service gracefully.
func (a *Archer) shutdown() error {

	// shut down the job chan to stop the process workers
	close(a.processChan)

	// shut down watcher channel if one exists
	if a.watcherChan != nil {
		close(a.watcherChan)
	}
	// sync and close the db
	if err := a.db.Sync(); err != nil {
		return err
	}
	if err := a.db.Close(); err != nil {
		return err
	}
	return nil
}

// addSample will add or update a sample
// in the Archer db.
// NOTE: this will nuke existing db entries
// and the user should check themselves
func (a *Archer) addSample(sample *api.SampleInfo) error {

	// lock the db for RW access
	a.Lock()
	defer a.Unlock()

	// marshal the sample and write it the db
	data, err := proto.Marshal(sample)
	if err != nil {
		return err
	}
	if err := a.db.Put([]byte(sample.GetSampleID()), data); err != nil {
		return err
	}
	return nil
}

// validateRequest will validate a service request.
func (a *Archer) validateRequest(request *api.ProcessRequest) error {

	// check input files exist
	if len(request.GetInputFASTQfiles()) == 0 {
		return fmt.Errorf("no FASTQ files provided")
	}
	for _, f := range request.GetInputFASTQfiles() {
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%v does not exist", f)
			}
			return fmt.Errorf("can't access %v", f)
		}
	}

	// check requested scheme is in the ARTIC manifest
	// and update the request the appropriate scheme tag
	// for this scheme
	schemeTag, err := amplicons.CheckManifest(a.manifest, request.GetScheme(), request.GetSchemeVersion())
	if err != nil {
		return err
	}
	request.Scheme = schemeTag

	// check that the current session has the requested amplicon set stored, or download it now
	if _, ok := a.ampliconCache[generateAmpliconSetID(request.GetScheme(), request.GetSchemeVersion())]; !ok {
		ampliconSet, err := amplicons.NewAmpliconSet(a.manifest, request.GetScheme(), request.GetSchemeVersion())
		if err != nil {
			return err
		}
		a.ampliconCache[generateAmpliconSetID(request.GetScheme(), request.GetSchemeVersion())] = ampliconSet
	}

	return nil
}

// generateAmpliconSetID is a helper function
// to generate a string id.
func generateAmpliconSetID(scheme string, version int32) string {
	return fmt.Sprintf("%v-%v", scheme, version)
}
