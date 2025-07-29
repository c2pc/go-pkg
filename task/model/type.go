package model

const ONLY_USERS_TASKS = "ONLY_USERS_TASKS"

const (
	Export     = "export"
	Import     = "import"
	MassUpdate = "mass-update"
	MassDelete = "mass-delete"
)

var Types = map[string]string{
	Export:     Export,
	Import:     Import,
	MassUpdate: MassUpdate,
	MassDelete: MassDelete,
}
