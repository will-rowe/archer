package service

import (
	log "github.com/sirupsen/logrus"
	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Watch is used to...
func (a *Archer) Watch(request *api.WatchRequest, stream api.Archer_WatchServer) error {
	log.Trace("watch request received...")

	// initialize the sample info array
	// samples := []*api.SampleInfo{}

	// loop until cancel requested
loop:
	for {
		select {
		case <-a.shutdownChan:
			log.Printf("stopping the watcher")
			break loop
		default:
			// get data/do processing

			// create a message and send it back
			resp := &api.WatchResponse{}
			if err := stream.Send(resp); err != nil {
				log.Printf("watch request failed to return a message: %v", err)
				return err
			}
		}
	}
	return nil
}
