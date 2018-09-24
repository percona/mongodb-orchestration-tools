package dcos

import (
	"testing"

	"github.com/percona/dcos-mongo-tools/internal/pod"
	"github.com/stretchr/testify/assert"
)

func TestDCOSTaskInterface(t *testing.T) {
	assert.Implements(t, (*pod.Task)(nil), &DCOSTask{})
}
