package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (job *Job) MakeImage() {
	state := job.GetVolumeState()
	if state != "detached" {
		fmt.Printf("ERROR volume not detached! Cannot image! Volume state: %s, Job state: %s\n", state, job.state.String())
		return
	}

	if job.snapshotID == "" {
		fmt.Println("ERROR no snapshot defined!")
		return
	}

	if job.snapshotState != "completed" {
		fmt.Println("ERROR no snapshot complete!")
		return
	}

	job.state = CreatingImage

	params := &ec2.CreateImageInput{
		VolumeId:   aws.String(job.volume),
		SnapshotId: aws.String(job.snapshot),
		DryRun:     aws.Bool(false),
	}
	resp, err := job.service.CreateSnapshot(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	job.snapshotID = *resp.SnapshotId

	fmt.Println("    > Snapshot ID: ", job.snapshotID)
}

func (job *Job) CheckImageState() string {
	if job.imageID == "" {
		fmt.Println("ERROR no image defined!")
		return ""
	}

	params := &ec2.DescribeImagesInput{
		DryRun: aws.Bool(false),
		ImageIds: []*string{
			aws.String(job.imageID),
		},
	}
	resp, err := job.service.DescribeImages(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return ""
	}

	state := *resp.Images[0].State
	if state == job.imageState {
		return state
	}

	job.imageState = state
	fmt.Printf("  > Image in state: %s\n", state)

	return state
}
