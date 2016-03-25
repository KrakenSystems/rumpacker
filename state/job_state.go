package state

type JobStatus int

//go:generate stringer -type=JobStatus
const (
	Initialised JobStatus = iota
	AMI_Detaching
	AMI_Snapshotting
	AMI_CreatingImage
	AMI_RegisteringImage
	AMI_Attaching
	Done
	Errored
	Attach_AWS_volume
)
