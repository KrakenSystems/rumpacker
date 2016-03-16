package rumpacker

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/KrakenSystems/ascalia-utils"
)

func (job *Job) CheckVolumeState() string {
	state := job.GetVolumeState()
	if state == job.volumeState {
		return state
	}

	job.volumeState = state
	fmt.Printf("\t> Volume in state: %s\n", state)

	return state
}

func (job *Job) GetVolumeState() string {
	params := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{
			aws.String(job.volume),
		},
	}
	resp, err := job.service.DescribeVolumes(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return ""
	}

	if len(resp.Volumes) != 1 {
		fmt.Println("ERROR: wrong number of returned volumes!")
		fmt.Println(resp)
		return ""
	}

	if len(resp.Volumes[0].Attachments) == 0 {
		return "detached"
	}

	if len(resp.Volumes[0].Attachments) > 1 {
		fmt.Println("ERROR: multiple attachments!")
		fmt.Println(resp)
		return "multiple"
	}

	return *resp.Volumes[0].Attachments[0].State
}

func (job *Job) ListVolumes() {
	params := &ec2.DescribeVolumesInput{
		MaxResults: aws.Int64(10),
	}
	resp, err := job.service.DescribeVolumes(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	fmt.Println("  > Number of volumes: ", len(resp.Volumes))
	for _, vol := range resp.Volumes {
		fmt.Println("    - Volume ID: ", *vol.VolumeId)
		for _, attach := range vol.Attachments {
			fmt.Printf("	- Attached to %s, status: %s\n", *attach.InstanceId, *attach.State)
		}
	}
}

func (job *Job) DetachVolume() {
	if job.state != Initialised {
		fmt.Printf("ERROR job not in state initialised! Cannot detach! State: %s\n", job.state.String())
		return
	}

	job.state = Detaching

	state := job.GetVolumeState()
	if state == "detached" {
		return
	}

	params := &ec2.DetachVolumeInput{
		VolumeId: aws.String(job.volume), // Required
		DryRun:   aws.Bool(false),
		Force:    aws.Bool(false),
	}
	fmt.Printf("\t> Detaching %s...\n", job.volume)

	_, err := job.service.DetachVolume(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
}

func (job *Job) AttachVolume() {
	state := job.GetVolumeState()
	if state != "detached" {
		fmt.Printf("ERROR volume not detached! Cannot attach! Volume state: %s, Job state: %s\n", state, job.state.String())
		return
	}

	job.state = Attaching

	params := &ec2.AttachVolumeInput{
		Device:     aws.String("/dev/sdf"),   // Required
		InstanceId: aws.String(job.instance), // Required
		VolumeId:   aws.String(job.volume),   // Required
		DryRun:     aws.Bool(false),
	}
	fmt.Printf("\t> Attaching %s...\n", job.volume)

	_, err := job.service.AttachVolume(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
}
