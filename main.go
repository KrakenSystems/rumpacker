package main

import (
	"fmt"
)

func main() {
	// Create an EC2 service object in the "us-west-2" region
	// Note that you can also configure your region globally by
	// exporting the AWS_REGION environment variable

	job := NewJob("i-f5d4d975", "vol-96207d49")

	job.ListInstances()
	job.ListVolumes()
	fmt.Println()

	job.Run()
	<-job.Done
}
