package healthcheck

type State int

const (
	StateOk State = iota
	StateFailed
)
