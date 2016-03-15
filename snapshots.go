package rumpacker

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (job *Job) MakeSnapshot() {
	state := job.GetVolumeState()
	if state != "detached" {
		fmt.Printf("ERROR volume not detached! Cannot snapshot! Volume state: %s, Job state: %s\n", state, job.state.String())
		return
	}

	job.state = Snapshotting

	params := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(job.volume), // Required
		Description: aws.String("some description lol"),
		DryRun:      aws.Bool(false),
	}
	resp, err := job.service.CreateSnapshot(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	job.snapshotID = *resp.SnapshotId

	fmt.Println("\t> Snapshot ID: ", job.snapshotID)
}

func (job *Job) CheckSnapshotState() string {
	if job.snapshotID == "" {
		fmt.Println("ERROR no snapshot defined!")
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
		fmt.Println(err.Error())
		return ""
	}

	state := *resp.Snapshots[0].State
	if state == job.snapshotState {
		return state
	}

	job.snapshotState = state
	fmt.Printf("\t> Snapshot in state: %s\n", state)

	return state
}
