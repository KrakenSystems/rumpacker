package rumpacker

type DatabaseJob struct {
	Id      int64 `db:"id"`
	UserId  int64 `db:"user_id"`
	RepoId  int64 `db:"repo_id"`
	RepoURL string

	TimeCreated int64
	TimeUpdated int64

	BuildType  string
	SourceType string

	Status JobStatus

	Log string
}
