package service

import (
	"context"
	"fmt"
	"os"

	"github.com/grailbio/bio/encoding/fastq"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/will-rowe/archer/pkg/api/v1"
)

// Process will begin processing for a sample.
func (a *Archer) Process(ctx context.Context, request *api.ProcessRequest) (*api.ProcessResponse, error) {
	log.Trace("process request received...")

	// check we have received a supported API request
	if err := a.checkAPI(request.GetApiVersion()); err != nil {
		return nil, err
	}

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
	if err := a.addSample(sampleInfo); err != nil {
		return nil, err
	}

	// add the sample to the processing queue
	a.processChan <- sampleInfo

	// create a response and return
	return &api.ProcessResponse{
		ApiVersion: a.version,
		Id:         sampleInfo.GetSampleID(),
	}, nil
}

// processWorker handles the actual work for Archer.
// This includes fastq checking, filtering, upload etc.
func (a *Archer) processWorker() {
	var read fastq.Read

	// collect requests
	for sample := range a.processChan {

		// make a copy of the amplicon set for this request
		as := a.ampliconCache[generateAmpliconSetID(sample.GetProcessRequest().GetScheme(), sample.GetProcessRequest().GetSchemeVersion())]

		// update the sample state, get the outfile handler and stats holder ready
		sample.State = api.State_RUNNING
		sample.ProcessStats = &api.SampleStats{
			TotalReads:       0,
			KeptReads:        0,
			AmpliconCoverage: make(map[string]int32),
			MeanAmpliconSize: int32(as.GetMeanSize()),
		}
		for amplicon := range *as {
			sample.ProcessStats.AmpliconCoverage[amplicon] = 0
		}
		lengthRange := int32(lengthThreshold * float64(sample.ProcessStats.MeanAmpliconSize))
		sample.ProcessStats.LengthMax = sample.ProcessStats.MeanAmpliconSize + lengthRange
		sample.ProcessStats.LengthMin = sample.ProcessStats.MeanAmpliconSize - lengthRange

		// open each FASTQ file for the sample
		for _, file := range sample.GetProcessRequest().GetInputFASTQfiles() {
			fh, err := os.Open(file)
			if checkError(sample, err) {
				continue
			}
			defer fh.Close()
			faScanner := fastq.NewScanner(fh, fastq.All)

			// filter reads against amplicons
			for faScanner.Scan(&read) {
				sample.ProcessStats.TotalReads++

				// length filter
				if len(read.Seq) < int(sample.ProcessStats.GetLengthMin()) || len(read.Seq) > int(sample.ProcessStats.GetLengthMax()) {
					continue
				}

				// filter against amplicons
				topHit, score, err := as.GetTopHit([]byte(read.Seq))
				if checkError(sample, err) {
					continue
				}
				if score < jaccardThreshold {
					continue
				}
				sample.ProcessStats.AmpliconCoverage[topHit]++

				// keep read, write to archive for upload
				sample.ProcessStats.KeptReads++
			}
		}

		// check for errors during fastq filtering
		// upload (TODO: decide if upload continues if errors found)

		// update status
		sample.State = api.State_SUCCESS

		// write back to db
		if err := a.addSample(sample); err != nil {
			panic(err)
		}

		// let a watcher know if needed
		if a.watcherChan != nil {
			a.watcherChan <- sample
		}
	}
}
