package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	log "github.com/sirupsen/logrus"
	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Process will begin processing for a sample.
func (a *Archer) Process(ctx context.Context, request *api.ProcessRequest) (*api.ProcessResponse, error) {
	log.Trace("process request received...")

	// check we have received a supported API request
	if err := a.checkAPI(request.GetApiVersion()); err != nil {
		return nil, err
	}

	// lock the db for RW access
	a.Lock()
	defer a.Unlock()

	// check we don't already have a request for this sample
	// TODO: we could make our lives harder by allowing samples to be repeated/edited etc.
	if a.db.Has([]byte(request.GetSampleID())) {
		return nil, status.Errorf(
			codes.AlreadyExists,
			fmt.Sprintf("duplicate sample can't be added to the database (%s)", request.GetSampleID()),
		)
	}

	// validate the request
	if err := a.validateRequest(request); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("request failed validation: %v", err),
		)
	}

	// create the sample info for archer
	sampleInfo, err := NewSample(SetID(request.GetSampleID()), SetRequest(request))
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("could not extract sample information: %v", err),
		)
	}

	// add the sample info to the db
	data, err := proto.Marshal(sampleInfo)
	if err != nil {
		return nil, err
	}
	if err := a.db.Put([]byte(sampleInfo.GetSampleID()), data); err != nil {
		return nil, err
	}

	// pass the sample to the processing queue
	// TODO...

	// create a response and return
	return &api.ProcessResponse{
		ApiVersion: a.version,
		Id:         sampleInfo.GetSampleID(),
	}, nil
}
