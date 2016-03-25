package rumpacker

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
	CloningRepo
	Started
	BuildingRump
	BuildingISO
	BuildingAMI
	Attach_AWS_volume
	Building_Jekyll
	Building_Clay
	Packing_static_web
	Generating_ISO_file
)
