package watcher

import (
	"testing"

	"github.com/percona/dcos-mongo-tools/internal/api"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
)

func TestGetScaledDownMembers(t *testing.T) {
	rs := replset.New(nil, "test")
	rs.UpdateMember(&replset.Mongod{
		Host:    "scaled-down",
		Port:    27017,
		PodName: "testPod",
	})

	w := &Watcher{
		activePods: &api.Pods{"testPod"},
		replset:    rs,
		state: &replset.State{
			Config: &rsConfig.Config{
				Members: []*rsConfig.Member{{
					Host: "scaled-down:27017",
				}},
			},
			Status: &rsStatus.Status{
				Members: []*rsStatus.Member{{
					Name:  "scaled-down:27017",
					State: rsStatus.MemberStatePrimary,
				}},
			},
		},
	}

	// test empty result (no down hosts and pod exists)
	assert.Len(t, w.getScaledDownMembers(), 0)

	// test empty result (1 down host but pod still exists)
	w.state.Status.Members[0].State = rsStatus.MemberStateDown
	assert.Len(t, w.getScaledDownMembers(), 0)

	// test scaled down (1 down host AND pod doesnt exist)
	w.activePods = &api.Pods{}
	assert.Len(t, w.getScaledDownMembers(), 1)
}
