package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	api "github.com/will-rowe/archer/pkg/api/v1"
	mock "github.com/will-rowe/archer/pkg/mock"
)

// TestArcher_Process will test the implementation of the
// Process rpc by Archer.
func TestArcher_Process(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockArcherClient(ctrl)

	// create a request
	req := &api.ProcessRequest{ApiVersion: apiVersion, InputFASTQfiles: []string{"./some/dir"}, Endpoint: "CLIMB"}

	// run the mock
	mockClient.EXPECT().Process(
		gomock.Any(),
		req,
	).Times(1).Return(&api.ProcessResponse{ApiVersion: apiVersion, Id: "archer-id"}, nil)
	res, err := mockClient.Process(context.Background(), req)

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.ApiVersion, apiVersion)
}
