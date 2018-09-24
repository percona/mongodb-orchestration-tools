package k8s

import (
	"testing"

	"github.com/percona/dcos-mongo-tools/internal/pod"
	"github.com/stretchr/testify/assert"
)

func TestK8STaskInterface(t *testing.T) {
	assert.Implements(t, (*pod.Task)(nil), &K8STask{})
}
