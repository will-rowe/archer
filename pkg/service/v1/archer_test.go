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

// newParticipant is a helper function to
// create a populated Participant struct
// for use in the tests.
func newParticipant() *api.Participant {
	t := time.Now().In(time.UTC)
	dob, _ := ptypes.TimestampProto(t)
	return &api.Participant{
		Id:      "KFG-734",
		Dob:     dob,
		Phone:   "123",
		Address: "The moon",
	}
}

// TestRegistryService_Create will test the implementation of the
// Create rpc by the RegistryService.
func TestRegistryService_Create(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockRegistryServiceClient(ctrl)

	// get a dummy participant and create a request
	p := newParticipant()
	req := &api.CreateRequest{ApiVersion: apiVersion, Participant: p}

	// run the mock
	mockClient.EXPECT().Create(
		gomock.Any(),
		req,
	).Times(1).Return(&api.CreateResponse{ApiVersion: apiVersion, Created: true}, nil)
	res, err := mockClient.Create(context.Background(), &api.CreateRequest{ApiVersion: apiVersion, Participant: p})

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.Created, true)
	assert.Equal(t, res.ApiVersion, apiVersion)
}

// TestRegistryService_Retrieve will test the implementation of the
// Retrieve rpc by the RegistryService.
func TestRegistryService_Retrieve(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockRegistryServiceClient(ctrl)

	// get a dummy participant and create a request
	p := newParticipant()
	req := &api.RetrieveRequest{ApiVersion: apiVersion, Id: p.GetId()}

	// run the mock
	mockClient.EXPECT().Retrieve(
		gomock.Any(),
		req,
	).Times(1).Return(&api.RetrieveResponse{ApiVersion: apiVersion, Participant: p}, nil)
	res, err := mockClient.Retrieve(context.Background(), &api.RetrieveRequest{ApiVersion: apiVersion, Id: p.GetId()})

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.Participant.GetId(), p.GetId())
	assert.Equal(t, res.Participant.GetDob(), p.GetDob())
	assert.Equal(t, res.Participant.GetPhone(), p.GetPhone())
	assert.Equal(t, res.Participant.GetAddress(), p.GetAddress())
}

// TestRegistryService_Update will test the implementation of the
// Update rpc by the RegistryService.
func TestRegistryService_Update(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockRegistryServiceClient(ctrl)

	// get a dummy participant and create a request
	p := newParticipant()
	req := &api.UpdateRequest{ApiVersion: apiVersion, Participant: p}

	// run the mock
	mockClient.EXPECT().Update(
		gomock.Any(),
		req,
	).Times(1).Return(&api.UpdateResponse{ApiVersion: apiVersion, Updated: true}, nil)
	res, err := mockClient.Update(context.Background(), &api.UpdateRequest{ApiVersion: apiVersion, Participant: p})

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.Updated, true)
}

// TestRegistryService_Delete will test the implementation of the
// Delete rpc by the RegistryService.
func TestRegistryService_Delete(t *testing.T) {

	// setup go mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockRegistryServiceClient(ctrl)

	// get a dummy participant and create a request
	p := newParticipant()
	req := &api.DeleteRequest{ApiVersion: apiVersion, Id: p.GetId()}

	// run the mock
	mockClient.EXPECT().Delete(
		gomock.Any(),
		req,
	).Times(1).Return(&api.DeleteResponse{ApiVersion: apiVersion, Deleted: true}, nil)
	res, err := mockClient.Delete(context.Background(), &api.DeleteRequest{ApiVersion: apiVersion, Id: p.GetId()})

	// check the results
	assert.NilError(t, err)
	assert.Equal(t, res.Deleted, true)
}

// TestAPIversion will check that API version requests
// are handled appropriately.
func TestAPIversion(t *testing.T) {
	v1 := "1"
	v2 := "2"
	rs := registryService{version: v1}
	if err := rs.checkAPI(v1); err != nil {
		t.Fatal(err)
	}
	if err := rs.checkAPI(v2); err == nil {
		t.Fatal("unsupported API missed by service API check")
	}
}

// TestDB will perform some basic tests for
// db checking.
func TestDB(t *testing.T) {
	req := &api.CreateRequest{ApiVersion: apiVersion, Participant: newParticipant()}
	rs := NewRegistryService()
	if _, err := rs.Create(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if _, err := rs.Create(context.Background(), req); err == nil {
		t.Fatal("duplicate participant added to db")
	}
	req2 := &api.DeleteRequest{ApiVersion: apiVersion, Id: "fake id"}
	if _, err := rs.Delete(context.Background(), req2); err == nil {
		t.Fatal("non-existent participant removed from db")
	}
}
