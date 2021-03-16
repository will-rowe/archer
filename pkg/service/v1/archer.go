// Package service implements the Archer service API.
package service

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/prologic/bitcask"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/will-rowe/archer/pkg/amplicons"
	api "github.com/will-rowe/archer/pkg/api/v1"
)

// useSync will run sync on every bit cask transaction, improving stability at the expense of time
const useSync = true

// apiVersion sets the API version to use
const apiVersion = "1"

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

	// shutdown is a signal to gracefully shutdown long running service processes (e.g. watch)
	shutdownChan chan struct{}

	// db is a key-value store for recording sample info
	db *bitcask.Bitcask

	// manifest is the ARTIC primer scheme manifest
	manifest *api.Manifest

	// ampliconCache is an in-memory cache of the schemes used for the current session
	ampliconCache map[string]*amplicons.AmpliconSet
}

// SetDb is an option setter for the NewArcher constructor
// that opens a db at the specified path and sets the
// appropriate field of the Archer struct.
func SetDb(dbPath string) ArcherOption {
	return func(x *Archer) error {

		// open/create the db
		db, err := bitcask.Open(dbPath, bitcask.WithSync(useSync))
		if err != nil {
			return err
		}

		// attach it to the Archer instance
		x.db = db
		return nil
	}
}

// SetManifest is an option setter for the NewArcher constructor
// that downloads and opens a manifest and sets the appropriate
// field of the Archer struct.
func SetManifest(manifestURL string) ArcherOption {
	return func(x *Archer) error {

		// download the manifest and unpack
		manifest, err := amplicons.GetManifest(manifestURL)
		if err != nil {
			return err
		}

		// attach it to the Archer instance
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
		shutdownChan:  make(chan struct{}),
		ampliconCache: make(map[string]*amplicons.AmpliconSet),
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

	// signal to any watch func calls that it's time to stop
	close(a.shutdownChan)

	// sync and close the db
	if err := a.db.Sync(); err != nil {
		return err
	}
	if err := a.db.Close(); err != nil {
		return err
	}
	return nil
}

// validateRequest will validate a service request.
func (a *Archer) validateRequest(request *api.ProcessRequest) error {

	// TODO: use type switch to validate more than just Process requests

	// check input files exist
	log.Trace("checking input files")
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
	log.Trace("checking manifest")
	if schemeTag, err := amplicons.CheckManifest(a.manifest, request.GetScheme(), request.GetSchemeVersion()); err != nil {
		return err
	} else {
		request.Scheme = schemeTag
	}

	// check that the current session has the requested amplicon set stored, or download it now
	key := fmt.Sprintf("%v-%v", request.GetScheme(), request.GetSchemeVersion())
	if _, ok := a.ampliconCache[key]; !ok {
		ampliconSet, err := amplicons.NewAmpliconSet(a.manifest, request.GetScheme(), request.GetSchemeVersion())
		if err != nil {
			return err
		}
		a.ampliconCache[key] = ampliconSet
	}

	// TODO: other checks (e.g. api endpoint)

	return nil
}
