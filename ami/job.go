package ami

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/KrakenSystems/ascalia-utils"
)

type Job struct {
	service     *ec2.EC2
	volume      string
	volumeState string

	instance string

	snapshotID    string
	snapshotState string

	imageName  string
	imageID    string
	imageState string

	kernelID string

	state JobStatus

	Wait        chan JobStatus
	StateChange chan JobStatus

	log chan string
}

func NewJob(instance string, volume string, kernelID string, log chan string) *Job {
	return &Job{
		service:     ec2.New(session.New(), &aws.Config{Region: aws.String("us-east-1")}),
		volume:      volume,
		instance:    instance,
		kernelID:    kernelID,
		Wait:        make(chan JobStatus),
		StateChange: make(chan JobStatus),
		log:         log,
	}
}

func (job *Job) SetState(state JobStatus) {
	job.state = state
	job.StateChange <- state
}

func (job *Job) GetState() JobStatus {
	return job.state
}

func (job *Job) GetImageID() string {
	return job.imageID
}

func (job *Job) SetImageID(img string) {
	job.imageID = img
}
