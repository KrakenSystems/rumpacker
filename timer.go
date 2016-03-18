package rumpacker

import (
	"fmt"
	"time"

	. "github.com/KrakenSystems/ascalia-utils"
)

func (job *Job) Run() {
	ticker := time.NewTicker(900 * time.Millisecond)

	prevState = -1

	go func() {
		for {
			<-ticker.C

			job.checkState()

			if job.state == Done {
				job.log <- "*** Job done! ***"
				break
			}
		}

		job.Done <- 1
	}()
}

var prevState JobStatus

func (job *Job) checkState() {

	if prevState != job.state {
		prevState = job.state
		job.log <- fmt.Sprintf("Job state: %s", job.state.String())
	}

	success := true

	switch job.state {

	case Initialised:
		success = job.DetachVolume()

	case AMI_Detaching:
		if job.CheckVolumeState() == "detached" {
			success = job.MakeSnapshot()
		}

	case AMI_Snapshotting:
		if job.CheckSnapshotState() == "completed" {
			success = job.RegisterImage()
		}

	case AMI_CreatingImage:
		if job.CheckImageState() == "available" {
			success = job.AttachVolume()
		}

	case ISOPrepareAMIVolume:
		success = job.AttachVolume()

	case AMI_Attaching:
		if job.CheckVolumeState() == "attached" {
			job.state = Done
		}
	}

	if !success {
		job.state = Done
		job.dbJob.SetStatus(Errored)
	}
}
