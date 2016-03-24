# rumpacker
Rumprun unikernel packing tool, currently just making an AMI image of a given instance/volume on AWS. You need to prepare the binary and data.iso on the machine itself yourself.

Coming soon: 

- actually preparing the Rumprun unikernel on the AWS instance
- documentation :)
- tests

# AWS credentials

AWS Go SDK is used to do work, so you need to give it proper AWS IAM credentials.

Create `~/.aws/credentials` and put following content inside:
```
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

# Minimal example

```golang
import (
    "github.com/ascaliaio/rumpacker/ami"
    . "github.com/ascaliaio/rumpacker/state"
)

func PackMyKernel(logChan chan string) bool {

	// logChan channel is a simple string channel where progress log will be sent to
	// together with any system output (shell commands & output)
    ami_job := ami.NewJob("i-123445", "vol-12321", "aki-1332342", logChan)

	// this triggers the job running, allows you to modify the job prior to running it
    ami_job.Run()

	// this is a blocking call, similar to sync.WaitGroup.Wait
    jobResult := ami_job.WaitJob()

	// jobResult is either Rumpacker_Errored or Rumpacker_Done

    if jobResult == Rumpacker_Errored {
        return false
    }

	return true
}
```
