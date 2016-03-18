package rumpacker

import (
	"fmt"
	"time"

	. "github.com/KrakenSystems/ascalia-utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	_ "github.com/KrakenSystems/ascalia-utils"
)

func (job *Job) MakeImage() bool {

	job.dbJob.SetStatus(AMI_CreatingImage)

	state := job.GetVolumeState()
	if state != "detached" {
		fmt.Printf("ERROR volume not detached! Cannot image! Volume state: %s, Job state: %s\n", state, job.state.String())
		return false
	}

	if job.snapshotID == "" {
		fmt.Println("ERROR no snapshot defined!")
		return false
	}

	if job.snapshotState != "completed" {
		fmt.Println("ERROR no snapshot complete!")
		return false
	}

	job.state = AMI_CreatingImage

	job.imageName = fmt.Sprintf("Image %d", time.Now().Unix())

	params := &ec2.CreateImageInput{
		InstanceId: aws.String(job.instance),  // Required
		Name:       aws.String(job.imageName), // Required
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{ // Required
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(false),
					SnapshotId:          aws.String(job.snapshotID),
					VolumeSize:          aws.Int64(1),
					VolumeType:          aws.String("standard"),
				},
			},
		},
		DryRun:   aws.Bool(false),
		NoReboot: aws.Bool(true),
	}

	resp, err := job.service.CreateImage(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return false
	}

	job.imageID = *resp.ImageId

	fmt.Println("\t> Image ID: ", job.imageID)
	return true
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
	fmt.Printf("\t> Image in state: %s\n", state)

	return state
}

func (job *Job) RegisterImage() bool {
	job.dbJob.SetStatus(AMI_RegisteringImage)

	if job.snapshotID == "" {
		fmt.Println("ERROR no snapshot defined!")
		return false
	}

	if job.snapshotState != "completed" {
		fmt.Println("ERROR no snapshot complete!")
		return false
	}

	job.imageName = fmt.Sprintf("Image %d", time.Now().Unix())

	params := &ec2.RegisterImageInput{
		Name:         aws.String(job.imageName), // Required
		Architecture: aws.String("x86_64"),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{ // Required
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					SnapshotId:          aws.String(job.snapshotID),
					VolumeSize:          aws.Int64(1),
					VolumeType:          aws.String("gp2"),
				},
			},
		},
		Description:        aws.String("String"),
		DryRun:             aws.Bool(false),
		KernelId:           aws.String(job.kernelID),
		RootDeviceName:     aws.String("/dev/sda1"),
		VirtualizationType: aws.String("paravirtual"),
	}
	resp, err := job.service.RegisterImage(params)

	job.state = AMI_CreatingImage

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return false
	}

	job.imageID = *resp.ImageId

	fmt.Println("\t> Image ID: ", job.imageID)
	return true
}
