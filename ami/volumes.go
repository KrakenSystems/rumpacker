package ami

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/KrakenSystems/rumpacker/state"
)

func (job *Job) GetVolumeState() (string, error) {
	params := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{
			aws.String(job.volume),
		},
	}
	resp, err := job.service.DescribeVolumes(params)

	if err != nil {
		return "", err
	}

	if len(resp.Volumes) != 1 {
		return "", errors.New("ERROR: wrong number of returned volumes!")
	}

	var volumeState string

	if len(resp.Volumes[0].Attachments) == 0 {
		volumeState = "detached"
	} else if len(resp.Volumes[0].Attachments) > 1 {
		volumeState = "multiple"
	} else {
		volumeState = *resp.Volumes[0].Attachments[0].State

	}

	if volumeState == job.volumeState {
		return volumeState, nil
	}

	job.volumeState = volumeState
	job.log <- fmt.Sprintf("\t> Volume in state: %s", volumeState)

	return volumeState, nil
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

func (job *Job) DetachVolume() error {
	if job.state != Initialised {
		return errors.New(fmt.Sprintf("ERROR job not in state initialised! Cannot detach! State: %s", job.state.String()))
	}

	job.SetState(AMI_Detaching)

	volumeState, err := job.GetVolumeState()
	if err != nil {
		return err
	} else if volumeState == "detached" {
		return nil
	}

	params := &ec2.DetachVolumeInput{
		VolumeId: aws.String(job.volume), // Required
		DryRun:   aws.Bool(false),
		Force:    aws.Bool(false),
	}
	job.log <- fmt.Sprintf("\t> Detaching %s...", job.volume)

	_, err = job.service.DetachVolume(params)

	if err != nil {
		return err
	}

	return nil
}

func (job *Job) AttachVolume() error {
	job.SetState(AMI_Attaching)

	state, err := job.GetVolumeState()
	if err != nil {
		return err
	} else if state != "detached" {
		return errors.New(fmt.Sprintf("ERROR volume not detached! Cannot attach! Volume state: %s, Job state: %s", state, job.state.String()))
	}

	params := &ec2.AttachVolumeInput{
		Device:     aws.String("/dev/sdf"),   // Required
		InstanceId: aws.String(job.instance), // Required
		VolumeId:   aws.String(job.volume),   // Required
		DryRun:     aws.Bool(false),
	}
	job.log <- fmt.Sprintf("\t> Attaching %s...", job.volume)

	_, err = job.service.AttachVolume(params)

	if err != nil {
		return err
	}

	return nil
}
