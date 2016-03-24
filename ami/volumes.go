package ami

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
	job.log <- fmt.Sprintf("\t> Volume in state: %s", state)

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
		job.log <- err.Error()
		return ""
	}

	if len(resp.Volumes) != 1 {
		job.log <- "ERROR: wrong number of returned volumes!"
		return ""
	}

	if len(resp.Volumes[0].Attachments) == 0 {
		return "detached"
	}

	if len(resp.Volumes[0].Attachments) > 1 {
		job.log <- "ERROR: multiple attachments!"
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

func (job *Job) DetachVolume() bool {
	if job.state != Initialised {
		job.log <- fmt.Sprintf("ERROR job not in state initialised! Cannot detach! State: %s", job.state.String())
		return false
	}

	job.dbJob.SetStatus(AMI_Detaching)
	job.state = AMI_Detaching

	state := job.GetVolumeState()
	if state == "detached" {
		return true
	}

	params := &ec2.DetachVolumeInput{
		VolumeId: aws.String(job.volume), // Required
		DryRun:   aws.Bool(false),
		Force:    aws.Bool(false),
	}
	job.log <- fmt.Sprintf("\t> Detaching %s...", job.volume)

	_, err := job.service.DetachVolume(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		job.log <- err.Error()
		return false
	}

	return true
}

func (job *Job) AttachVolume() bool {
	job.dbJob.SetStatus(AMI_Attaching)
	job.state = AMI_Attaching

	state := job.GetVolumeState()
	if state != "detached" {
		job.log <- fmt.Sprintf("ERROR volume not detached! Cannot attach! Volume state: %s, Job state: %s", state, job.state.String())
		return false
	}

	params := &ec2.AttachVolumeInput{
		Device:     aws.String("/dev/sdf"),   // Required
		InstanceId: aws.String(job.instance), // Required
		VolumeId:   aws.String(job.volume),   // Required
		DryRun:     aws.Bool(false),
	}
	job.log <- fmt.Sprintf("\t> Attaching %s...", job.volume)

	_, err := job.service.AttachVolume(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return false
	}

	return true
}
