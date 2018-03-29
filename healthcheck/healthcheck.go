package healthcheck

type ExitCode int

const (
	ExitCodeOk     ExitCode = 0
	ExitCodeFailed ExitCode = 1
)
