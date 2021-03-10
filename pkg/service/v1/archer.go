// Package service implements the Archer service API.
package service

import (
	"context"
	"sync"

	api "github.com/will-rowe/archer/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	apiVersion = "1"
)

// archer is an implementation
// of the v1.ArcherServer.
type archer struct {

	// version of API implemented by the server
	version string

	// db is the in-memory db
	db map[string]api.SampleInfo

	// db lock
	sync.RWMutex
}

// NewArcher creates the Archer service.
func NewArcher() api.ArcherServer {
	return &archer{
		version: apiVersion,
		db:      make(map[string]api.SampleInfo),
	}
}

// checkAPI checks if requested API version is supported
// by the server.
func (a *archer) checkAPI(requestedAPI string) error {
	if a.version != requestedAPI {
		return status.Errorf(codes.Unimplemented,
			"unsupported API version requested: current Archer service implements version '%s', but version '%s' was requested", a.version, requestedAPI)
	}
	return nil
}

// Start will begin processing for a sample.
func (a *archer) Start(ctx context.Context, request *api.StartRequest) (*api.StartResponse, error) {

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
