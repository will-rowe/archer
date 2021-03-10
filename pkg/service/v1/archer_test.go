package service

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	api "github.com/will-rowe/archer/pkg/api/v1"
	mock "github.com/will-rowe/archer/pkg/mock"
)

var (
	dbLocation string = "./tmp"
)

// cleanUp is called to remove the database
// after testing completes
func cleanUp() error {
	return os.RemoveAll(dbLocation)
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
	a, err := NewArcher(dbLocation)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.checkAPI(v1); err != nil {
		t.Fatal(err)
	}
	if err := a.checkAPI(v2); err == nil {
		t.Fatal("unsupported API missed by service API check")
	}
	if err := cleanUp(); err != nil {
		t.Fatal(err)
	}
}
