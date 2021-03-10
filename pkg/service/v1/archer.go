// Package service implements the Archer service API.
package service

import (
	"context"
	"log"
	"sync"

	api "github.com/will-rowe/archer/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	apiVersion = "1"
)

// Archer is an implementation
// of the v1.ArcherServer.
type Archer struct {

	// db lock
	sync.RWMutex

	// version of API implemented by the server
	version string

	// TODO: db is the in-memory db and will be replaced soon...
	db map[string]api.SampleInfo

	// shutdown is a signal to gracefully shutdown long running service processes (e.g. watch)
	shutdown chan struct{}
}

// NewArcher creates the Archer service.
func NewArcher() *Archer {
	return &Archer{
		version:  apiVersion,
		db:       make(map[string]api.SampleInfo),
		shutdown: make(chan struct{}),
	}
}

// Shutdown will stop the Archer service gracefully.
func (a *Archer) Shutdown() error {

	// TODO: db flush etc.

	// signal to any watch func calls that it's time to stop
	close(a.shutdown)

	return nil
}

// Start will begin processing for a sample.
func (a *Archer) Start(ctx context.Context, request *api.StartRequest) (*api.StartResponse, error) {

	// check we have received a supported API request
	if err := a.checkAPI(request.GetApiVersion()); err != nil {
		return nil, err
	}

	// lock the db for RW access
	a.Lock()
	defer a.Unlock()

	// check if entry already exists for provided reference number
	//if _, ok := rs.db[request.GetParticipant().GetId()]; ok {
	//	return nil, status.Errorf(codes.AlreadyExists,
	//		"reference number in use: participant already exists in the registry for %v", request.GetParticipant().GetId())
	//}

	// TODO: validate the provided participant details

	// add the participant as an entry in the registry db
	//rs.db[request.GetParticipant().GetId()] = *request.GetParticipant()

	// create a response and return
	return &api.StartResponse{
		ApiVersion: a.version,
		Id:         "blam",
	}, nil
}

// Cancel is used to...
func (a *Archer) Cancel(ctx context.Context, request *api.CancelRequest) (*api.CancelResponse, error) {

	return nil, nil
}

// Watch is used to...
func (a *Archer) Watch(request *api.WatchRequest, stream api.Archer_WatchServer) error {
	log.Println("watch called...")

	// initialize the sample info array
	// samples := []*api.SampleInfo{}

	// loop until cancel requested
loop:
	for {
		select {
		case <-a.shutdown:
			break loop
		default:
			// get data/do processing

			// create a message
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
