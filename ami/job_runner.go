package ami

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

			if job.state == Done || job.state == Errored {
				job.log <- "*** Job done! ***"
				break
			}
		}

		job.Wait <- job.state
	}()
}

var prevState JobStatus

func (job *Job) checkState() {

	if prevState != job.state {
		prevState = job.state
		job.log <- fmt.Sprintf("Job state: %s", job.state.String())
	}

	var err error

	switch job.state {

	case Initialised:
		err = job.DetachVolume()

	case AMI_Detaching:
		var state string
		state, err = job.GetVolumeState()
		if state == "detached" {
			err = job.MakeSnapshot()
		}

	case AMI_Snapshotting:
		var state string
		state, err = job.GetSnapshotState()
		if state == "completed" {
			err = job.RegisterImage()
		}

	case AMI_CreatingImage:
		// Wait for image to become available, after that, make it public
		var state string
		state, err = job.GetImageState()
		if state == "available" {
			err = job.ImageSetPublic()
		}

	case Attach_AWS_volume:
		// Intermediary state used only when AWS Volume is not available as a prerequisite
		err = job.AttachVolume()

	case AMI_Attaching:
		// Wait for Volume to be attached again, then we're done
		var state string
		state, err = job.GetVolumeState()
		if state == "attached" {
			job.SetState(Done)
		}
	}

	if err != nil {
		job.log <- err.Error()
		job.SetState(Errored)
	}
}
