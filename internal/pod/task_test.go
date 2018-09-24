package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskInterface(t *testing.T) {
	assert.Implements(t, (*Task)(nil), &DCOSTask{})
	assert.Implements(t, (*Task)(nil), &K8STask{})
}
