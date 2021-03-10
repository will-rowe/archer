package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"gotest.tools/assert"

	api "github.com/will-rowe/archer/pkg/api/v1"
	mock "github.com/will-rowe/archer/pkg/mock"
)

// newSampleInfo is a helper function to
// create a populated SampleInfo struct
// for use in the tests.
func newSampleInfo() *api.SampleInfo {
	t := time.Now().In(time.UTC)
	startTime, _ := ptypes.TimestampProto(t)
	return &api.SampleInfo{
		Id:        "sampleXYZ",
		StartTime: startTime,
	}
}

// TestArcher_Start will test the implementation of the
// Start rpc by Archer.
func TestArcher_Start(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockArcherClient(ctrl)

	// create a request
	req := &api.StartRequest{ApiVersion: apiVersion, InputReadsDirectories: []string{"./some/dir"}, Endpoint: "CLIMB"}

	// run the mock
	mockClient.EXPECT().Start(
		gomock.Any(),
		req,
	).Times(1).Return(&api.StartResponse{ApiVersion: apiVersion, Id: "archer-id"}, nil)
	res, err := mockClient.Start(context.Background(), req)

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.ApiVersion, apiVersion)
}

// TestAPIversion will check that API version requests
// are handled appropriately.
func TestAPIversion(t *testing.T) {
	v1 := "1"
	v2 := "2"
	a := NewArcher()
	if err := a.checkAPI(v1); err != nil {
		t.Fatal(err)
	}
	if err := a.checkAPI(v2); err == nil {
		t.Fatal("unsupported API missed by service API check")
	}
}
