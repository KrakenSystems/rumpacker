package main

import (
	"fmt"
	"time"
)

func (job *Job) Run() {
	ticker := time.NewTicker(900 * time.Millisecond)

	prevState = -1

	go func() {
		for {
			<-ticker.C

			job.checkState()

			if job.state == Done {
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
		fmt.Printf("Job state: %s\n", job.state.String())
	}

	switch job.state {

	case Initialised:
		job.DetachVolume()

	case Detaching:
		if job.CheckVolumeState() == "detached" {
			job.MakeSnapshot()
		}

	case Snapshotting:
		if job.CheckSnapshotState() == "completed" {
			job.MakeImage()
		}

	case CreatingImage:
		if job.CheckImageState() == "available" {
			job.AttachVolume()
		}

	case Attaching:
		if job.CheckVolumeState() == "attached" {
			job.state = Done
		}
	}
}
