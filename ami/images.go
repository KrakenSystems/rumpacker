package ami

import (
	"errors"
	"fmt"
	"time"

	. "github.com/KrakenSystems/ascalia-utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	_ "github.com/KrakenSystems/ascalia-utils"
)

func (job *Job) GetImageState() (string, error) {
	if job.imageID == "" {
		job.log <- "ERROR no image defined!"
		return "", nil
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
		return "", err
	}

	state := *resp.Images[0].State
	if state == job.imageState {
		return state, nil
	}

	job.imageState = state
	job.log <- fmt.Sprintf("> Image in state: %s", state)

	return state, nil
}

func (job *Job) RegisterImage() error {
	job.dbJob.SetStatus(AMI_RegisteringImage)
	job.state = AMI_RegisteringImage

	if job.snapshotID == "" {
		return errors.New("ERROR no snapshot defined!")
	}

	if job.snapshotState != "completed" {
		return errors.New("ERROR no snapshot complete!")
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
		return err
	}

	job.imageID = *resp.ImageId

	job.log <- fmt.Sprintf("\t> Image ID: %s", job.imageID)

	job.SetState(AMI_CreatingImage)

	return nil
}

func (job *Job) ImageSetPublic() error {
	if job.imageID == "" {
		return errors.New("ERROR no image defined!")
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
	_, err := job.service.ModifyImageAttribute(params)

	if err != nil {
		return err
	}

	job.SetState(Attach_AWS_volume)

	return nil
}
