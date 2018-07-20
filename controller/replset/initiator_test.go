package replset

import (
	gotesting "testing"

	//"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	//"github.com/percona/dcos-mongo-tools/common/logger"
	//"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/stretchr/testify/assert"
)

var (
	testInitiator *Initiator
	testConfig    = &controller.Config{
		SSL: &db.SSLConfig{},
	}
)

func TestNewInitiator(t *gotesting.T) {
	testInitiator = NewInitiator(testConfig)
	assert.NotNil(t, testInitiator)
}
