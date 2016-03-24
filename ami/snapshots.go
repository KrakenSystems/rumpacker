package ami

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/KrakenSystems/ascalia-utils"
)

func (job *Job) MakeSnapshot() bool {
	job.dbJob.SetStatus(AMI_Snapshotting)

	state := job.GetVolumeState()
	if state != "detached" {
		job.log <- fmt.Sprintf("ERROR volume not detached! Cannot snapshot! Volume state: %s, Job state: %s", state, job.state.String())
		return false
	}
	job.state = AMI_Snapshotting

	params := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(job.volume), // Required
		Description: aws.String("some description lol"),
		DryRun:      aws.Bool(false),
	}
	resp, err := job.service.CreateSnapshot(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		job.log <- err.Error()
		return false
	}

	job.snapshotID = *resp.SnapshotId

	job.log <- fmt.Sprintf("\t> Snapshot ID: %s", job.snapshotID)
	return true
}

func (job *Job) CheckSnapshotState() string {
	if job.snapshotID == "" {
		job.log <- "ERROR no snapshot defined!"
		return ""
	}

	params := &ec2.DescribeSnapshotsInput{
		DryRun: aws.Bool(false),
		SnapshotIds: []*string{
			aws.String(job.snapshotID), // Required
		},
	}
	resp, err := job.service.DescribeSnapshots(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		job.log <- err.Error()
		return ""
	}

	state := *resp.Snapshots[0].State
	if state == job.snapshotState {
		return state
	}

	job.snapshotState = state
	job.log <- fmt.Sprintf("\t> Snapshot in state: %s", state)

	return state
}
