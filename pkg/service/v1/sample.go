package service

import (
	"errors"

	"github.com/golang/protobuf/ptypes"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// SampleOption is a wrapper struct used to pass functional
// options to the Sample constructor.
type SampleOption func(sample *api.SampleInfo) error

// SetID is an option setter for the NewSample constructor
// that sets the ID field of a SampleInfo struct.
func SetID(id string) SampleOption {
	return func(x *api.SampleInfo) error {
		if len(id) == 0 {
			return errors.New("sample requires an ID to be provided")
		}
		x.Id = id
		return nil
	}
}

// SetRequest is an option setter for the NewSample constructor
// that sets the request field of a SampleInfo struct.
func SetRequest(request *api.ProcessRequest) SampleOption {
	return func(x *api.SampleInfo) error {

		// TODO: validate the request?

		x.ProcessRequest = request
		return nil
	}
}

// NewSample will return an initialised
// SampleInfo struct with default values.
func NewSample(options ...SampleOption) (*api.SampleInfo, error) {

	// construct
	s := &api.SampleInfo{
		State:           api.State_STATE_RUNNING,
		Errors:          []string{},
		FilesDiscovered: 0,
		StartTime:       ptypes.TimestampNow(),
	}

	// set options
	for _, option := range options {
		err := option(s)
		if err != nil {
			return nil, err
		}
	}

	// return the initialised struct
	return s, nil
}
