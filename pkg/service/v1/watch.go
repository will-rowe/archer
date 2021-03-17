package service

import (
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Watch is used to...
func (a *Archer) Watch(request *api.WatchRequest, stream api.Archer_WatchServer) error {
	log.Trace("watch request received...")

	// set up a chan for the watcher to receive updates on
	a.watcherChan = make(chan *api.SampleInfo)

	// start with a loop over the db and send all finished samples
	sample := &api.SampleInfo{}
	response := &api.WatchResponse{
		ApiVersion: a.version,
		Samples:    []*api.SampleInfo{},
	}
	for key := range a.db.Keys() {
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
		log.Printf("watch request failed to return a message: %v", err)
		return err
	}

	// loop until cancel requested
loop:
	for {
		select {
		case <-a.shutdownChan:
			log.Printf("stopping the watcher")
			break loop
		default:
			sample := <-a.watcherChan
			resp := &api.WatchResponse{ApiVersion: a.version, Samples: []*api.SampleInfo{sample}}
			if err := stream.Send(resp); err != nil {
				log.Printf("watch request failed to return a message: %v", err)
				return err
			}
		}
	}
	return nil
}
