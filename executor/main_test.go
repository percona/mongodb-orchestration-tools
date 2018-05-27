package executor

import (
	"os"
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/executor/config"
)

var (
	testExecutor               *Executor
	testExecutorDaemon         Daemon
	testExecutorDaemonNodeType = config.NodeType("MockDaemon")
	testExecutorConfig         = &config.Config{
		NodeType: testExecutorDaemonNodeType,
	}
	testQuitChan = make(chan bool)
)

func TestMain(m *gotesting.M) {
	exit := m.Run()
	if testExecutorDaemon != nil {
		testExecutorDaemon.Kill()
	}
	os.Exit(exit)
}
