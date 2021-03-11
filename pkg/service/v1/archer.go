// Package service implements the Archer service API.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/prologic/bitcask"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// useSync will run sync on every bit cask transaction, improving stability at the expense of time
const useSync = true

// apiVersion sets the API version to use
const apiVersion = "1"

// Archer is an implementation
// of the v1.ArcherServer.
type Archer struct {

	// db lock
	sync.RWMutex

	// version of API implemented by the server
	version string

	// db is a key-value store for recording sample info
	db *bitcask.Bitcask

	// shutdown is a signal to gracefully shutdown long running service processes (e.g. watch)
	shutdownChan chan struct{}
}

// NewArcher creates the Archer server and returns
// it along with the shutdown method and any
// constructor error.
func NewArcher(dbPath string) (api.ArcherServer, func() error, error) {

	// open/create the db
	db, err := bitcask.Open(dbPath, bitcask.WithSync(useSync))
	if err != nil {
		return nil, nil, err
	}

	// create the service and return it
	a := &Archer{
		version:      apiVersion,
		db:           db,
		shutdownChan: make(chan struct{}),
	}

	return a, a.shutdown, nil
}

// Start will begin processing for a sample.
func (a *Archer) Start(ctx context.Context, request *api.StartRequest) (*api.StartResponse, error) {
	log.Trace("start request received...")

	// check we have received a supported API request
	if err := a.checkAPI(request.GetApiVersion()); err != nil {
		return nil, err
	}

	// lock the db for RW access
	a.Lock()
	defer a.Unlock()

	// check we don't already have a request for this sample
	// TODO: we could make our lives harder by allowing samples to be repeated
	if a.db.Has([]byte(request.GetId())) {
		return nil, fmt.Errorf("duplicate sample can't be added to the database (%s)", request.GetId())
	}

	// create the sample info for the request
	sampleInfo, err := NewSample(SetID(request.GetId()), SetRequest(request))
	if err != nil {
		return nil, err
	}

	// add the sample info to the db
	data, err := proto.Marshal(sampleInfo)
	if err != nil {
		return nil, err
	}
	if err := a.db.Put([]byte(sampleInfo.GetId()), data); err != nil {
		return nil, err
	}

	// pass the sample to the processing queue
	// TODO...

	// create a response and return
	return &api.StartResponse{
		ApiVersion: a.version,
		Id:         sampleInfo.GetId(),
	}, nil
}

// Cancel is used to...
func (a *Archer) Cancel(ctx context.Context, request *api.CancelRequest) (*api.CancelResponse, error) {

	return nil, nil
}

// Watch is used to...
func (a *Archer) Watch(request *api.WatchRequest, stream api.Archer_WatchServer) error {
	log.Trace("watch request received...")

	// initialize the sample info array
	// samples := []*api.SampleInfo{}

	// loop until cancel requested
loop:
	for {
		select {
		case <-a.shutdownChan:
			log.Printf("stopping the watcher")
			break loop
		default:
			// get data/do processing

			// create a message and send it back
			resp := &api.WatchResponse{}
			if err := stream.Send(resp); err != nil {
				log.Printf("watch request failed to return a message: %v", err)
				return err
			}
		}
	}
	return nil
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
