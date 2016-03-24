package ami

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
		job.log <- fmt.Sprintf("ERROR volume not detached! Cannot image! Volume state: %s, Job state: %s", state, job.state.String())
		return false
	}
	job.state = AMI_CreatingImage

	if job.snapshotID == "" {
		job.log <- "ERROR no snapshot defined!"
		return false
	}

	if job.snapshotState != "completed" {
		job.log <- "ERROR no snapshot complete!"
		return false
	}

	job.imageName = fmt.Sprintf("Image %d", time.Now().Unix())
	job.log <- fmt.Sprintf("AWS image name: %s", job.imageName)

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
		job.log <- err.Error()
		return false
	}

	job.imageID = *resp.ImageId

	job.log <- fmt.Sprintf("> Image ID: %s", job.imageID)
	return true
}

func (job *Job) CheckImageState() string {
	if job.imageID == "" {
		job.log <- "ERROR no image defined!"
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
		job.log <- err.Error()
		return ""
	}

	state := *resp.Images[0].State
	if state == job.imageState {
		return state
	}

	job.imageState = state
	job.log <- fmt.Sprintf("> Image in state: %s", state)

	return state
}

func (job *Job) RegisterImage() bool {
	job.dbJob.SetStatus(AMI_RegisteringImage)
	job.state = AMI_RegisteringImage

	if job.snapshotID == "" {
		job.log <- "ERROR no snapshot defined!"
		return false
	}

	if job.snapshotState != "completed" {
		job.log <- "ERROR no snapshot complete!"
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

	if err != nil {
		job.log <- err.Error()
		return false
	}

	job.imageID = *resp.ImageId

	job.log <- fmt.Sprintf("\t> Image ID: %s", job.imageID)

	job.SetState(AMI_CreatingImage)

	return true
}

func (job *Job) ImageSetPublic() bool {
	if job.imageID == "" {
		job.log <- "ERROR no image defined!"
		return false
	}

	params := &ec2.ModifyImageAttributeInput{
		ImageId: aws.String(job.imageID), // Required
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: []*ec2.LaunchPermission{
				{
					Group: aws.String("all"),
				},
			},
		},
		OperationType: aws.String("add"),
	}
	resp, err := job.service.ModifyImageAttribute(params)

	if err != nil {
		job.log <- err.Error()
		return false
	}

	job.log <- fmt.Sprintf("Response: %+v", resp)
	job.SetState(Attach_AWS_volume)

	return true
}
