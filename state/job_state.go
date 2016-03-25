package state

type RumpackerState int

//go:generate stringer -type=RumpackerState
const (
	Initialised RumpackerState = iota
	AMI_Detaching
	AMI_Snapshotting
	AMI_CreatingImage
	AMI_RegisteringImage
	AMI_Attaching
	Done
	Errored
	Attach_AWS_volume
)
