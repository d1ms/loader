package misc

// Worker is work
type Worker struct {
	Links  []string
	Status bool
	Id     int
}

// Manager is manage
type Manager struct {
	Workers []*Worker
}
