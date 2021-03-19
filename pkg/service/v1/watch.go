package service

import (
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Watch will respond to WatchRequests by streaming completed samples
// back to the user as they are marked done by Archer.
func (a *Archer) Watch(request *api.WatchRequest, stream api.Archer_WatchServer) error {
	log.Info("watch request received")

	// set up a chan for the watcher to receive updates on
	if a.watcherChan == nil {
		a.watcherChan = make(chan *api.SampleInfo)
	}

	// start with a loop over the db and send all finished samples (if requested)
	if request.GetSendFinished() {
		response := &api.WatchResponse{
			ApiVersion: a.version,
			Samples:    []*api.SampleInfo{},
		}
		for key := range a.db.Keys() {
			sample := &api.SampleInfo{}
			data, err := a.db.Get(key)
			if err != nil {
				return err
			}
			if err := proto.Unmarshal(data, sample); err != nil {
				return err
			}
			if sample.GetState() == api.State_SUCCESS {
				response.Samples = append(response.Samples, sample)
			}
		}
		if err := stream.Send(response); err != nil {
			log.Errorf("watch request failed to return a message: %v", err)
			return err
		}
	}

	// collect completed samples as they are finished by the Process workers
	errChan := make(chan error)
	go func() {
		for sample := range a.watcherChan {
			log.Infof("watcher received finished sample - %s", sample.GetSampleID())
			resp := &api.WatchResponse{ApiVersion: a.version, Samples: []*api.SampleInfo{sample}}
			if err := stream.Send(resp); err != nil {
				errChan <- err
			}
		}
	}()

	// wait for a shutdown or error
	for {
		select {
		case err := <-errChan:
			return err
		case <-stream.Context().Done():
			log.Info("closing watcher stream")
			close(errChan)
			return nil
		}
	}
}
