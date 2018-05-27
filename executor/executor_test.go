package executor

import (
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

type MockDaemon struct{}

func (d *MockDaemon) Name() string {
	return "MockDaemon"
}

func (d *MockDaemon) IsStarted() bool {
	return false
}

func (d *MockDaemon) Start() error {
	return nil
}

func (d *MockDaemon) Wait() {
}

func (d *MockDaemon) Kill() error {
	return nil
}

func TestExecutorNew(t *gotesting.T) {
	testExecutor = New(testExecutorConfig, &testQuitChan)
	assert.NotNil(t, testExecutor, ".New() should not return nil")
}

func TestExecutorRun(t *gotesting.T) {
	testExecutorDaemon = &MockDaemon{}
	err := testExecutor.Run(testExecutorDaemon)
	assert.NoError(t, err, ".Run() should not return an error")
}
