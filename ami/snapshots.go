package ami

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/KrakenSystems/rumpacker"
)

func (job *Job) MakeSnapshot() error {
	job.SetState(AMI_Snapshotting)

	volumeState, err := job.GetVolumeState()
	if err != nil {
		return err
	} else if volumeState != "detached" {
		return errors.New(fmt.Sprintf("ERROR volume not detached! Cannot snapshot! Volume state: %s, Job state: %s", volumeState, job.state.String()))
	}

	params := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(job.volume),
		Description: aws.String("snapshot description"),
		DryRun:      aws.Bool(false),
	}
	resp, err := job.service.CreateSnapshot(params)

	if err != nil {
		return err
	}

	job.snapshotID = *resp.SnapshotId

	job.log <- fmt.Sprintf("\t> Snapshot ID: %s", job.snapshotID)
	return nil
}

func (job *Job) GetSnapshotState() (string, error) {
	if job.snapshotID == "" {
		return "", errors.New("ERROR no snapshot defined!")
	}

	params := &ec2.DescribeSnapshotsInput{
		DryRun: aws.Bool(false),
		SnapshotIds: []*string{
			aws.String(job.snapshotID), // Required
		},
	}
	resp, err := job.service.DescribeSnapshots(params)

	if err != nil {
		return "", err
	}

	snapshotState := *resp.Snapshots[0].State
	if snapshotState == job.snapshotState {
		return snapshotState, nil
	}

	job.snapshotState = snapshotState
	job.log <- fmt.Sprintf("\t> Snapshot in state: %s", snapshotState)

	return snapshotState, nil
}
