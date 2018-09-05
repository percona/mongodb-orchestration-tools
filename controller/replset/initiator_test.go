package replset

import (
	"fmt"
	"os"
	"testing"

	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/percona/dcos-mongo-tools/controller/user"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/logger"
	"github.com/percona/dcos-mongo-tools/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

var (
	testSession   *mgo.Session
	testInitiator *Initiator
	testConfig    = &controller.Config{
		SSL: &db.SSLConfig{},
	}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)

	if testutils.Enabled() {
		var err error
		testSession, err = testutils.GetSession(testutils.MongodbPrimaryPort)
		if err != nil {
			fmt.Printf("Error getting session: %v", err)
			os.Exit(1)
		}
	}
	exit := m.Run()
	if testSession != nil {
		testSession.Close()
	}
	os.Exit(exit)
}

func TestNewInitiator(t *testing.T) {
	testInitiator = NewInitiator(testConfig)
	assert.NotNil(t, testInitiator)
}

func TestControllerReplsetInitiatorInitAdminUser(t *testing.T) {
	testutils.DoSkipTest(t)

	user.UserAdmin = &mgo.User{
		Username: "testAdmin",
		Password: "testAdminPassword",
		Roles: []mgo.Role{
			mgo.RoleUserAdminAny,
		},
	}
	assert.NoError(t, testInitiator.initAdminUser(testSession))
	assert.NoError(t, user.RemoveUser(testSession, user.UserAdmin.Username, "admin"))
}

func TestControllerReplsetInitiatorInitUsers(t *testing.T) {
	testutils.DoSkipTest(t)

	user.SystemUsers = []*mgo.User{
		{
			Username: "testUser",
			Password: "testUserPassword",
			Roles: []mgo.Role{
				mgo.RoleReadWrite,
			},
		},
	}
	assert.NoError(t, testInitiator.initUsers(testSession))
	assert.NoError(t, user.RemoveUser(testSession, user.SystemUsers[0].Username, "admin"))
}
