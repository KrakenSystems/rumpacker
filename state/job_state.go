package state

type RumpackerState int

//go:generate stringer -type=RumpackerState
const (
	Rumpacker_Initialised RumpackerState = iota
	AMI_Detaching
	AMI_Snapshotting
	AMI_CreatingImage
	AMI_RegisteringImage
	AMI_Attaching
	Rumpacker_Done
	Rumpacker_Errored
	Attach_AWS_volume
)
