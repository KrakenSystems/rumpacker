package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type JobStatus int

//go:generate stringer -type=JobStatus
const (
	Initialised JobStatus = iota
	Detaching
	Snapshotting
	CreatingImage
	Attaching
	Done
)

type Job struct {
	service     *ec2.EC2
	volume      string
	volumeState string

	instance string

	snapshotID    string
	snapshotState string

	imageID    string
	imageState string

	state JobStatus
	Done  chan int
}

func NewJob(instance string, volume string) *Job {
	return &Job{
		service:  ec2.New(session.New(), &aws.Config{Region: aws.String("us-east-1")}),
		volume:   volume,
		instance: instance,
		Done:     make(chan int),
	}
}

func (job *Job) GetState() {
	fmt.Printf("Job status: %s\n", job.state.String())
}
