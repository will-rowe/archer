package service

import (
	"errors"
	"fmt"

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
		x.SampleID = id
		return nil
	}
}

// SetRequest is an option setter for the NewSample constructor
// that sets the request field of a SampleInfo struct and
// populates some additional data.
func SetRequest(request *api.ProcessRequest) SampleOption {
	return func(x *api.SampleInfo) error {
		x.ProcessRequest = request
		x.FilesDiscovered = int32(len(request.GetInputFASTQfiles()))
		return nil
	}
}

// NewSample will return an initialised
// SampleInfo struct with default values.
func NewSample(options ...SampleOption) (*api.SampleInfo, error) {

	// construct
	s := &api.SampleInfo{
		SampleID:        "",
		ProcessRequest:  nil,
		State:           api.State_UNKNOWN,
		Errors:          []string{},
		FilesDiscovered: 0,
		StartTime:       ptypes.TimestampNow(),
	}

	// set options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	// return the initialised struct
	return s, nil
}

// GetAmpliconCoverage will check SampleStats and return
// the number of covered amplicons, the total number of
// amplicons and the mean coverage.
func GetAmpliconCoverage(sampleStats *api.SampleStats) (int, int, float64) {
	coveredAmplicons := 0
	totalAmplicons := len(sampleStats.AmpliconCoverage)
	meanCoverage := int32(0)
	for _, cov := range sampleStats.AmpliconCoverage {
		if cov != 0 {
			coveredAmplicons++
		}
		meanCoverage += cov
	}
	return coveredAmplicons, totalAmplicons, float64(meanCoverage) / float64(totalAmplicons)
}

// checkError will check an error, add it to the sample
// and update its state. True is returned if an error
// was received, False if error was nil.
func checkError(sample *api.SampleInfo, err error) bool {
	if err == nil {
		return false
	}
	sample.State = api.State_ERROR
	sample.Errors = append(sample.Errors, fmt.Sprintf("%s", err))
	return true
}
