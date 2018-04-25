package healthcheck

import (
	"testing"

	"github.com/timvaillancourt/go-mongodb-replset/status"
)

var (
	testMember = &status.Member{
		Id:       0,
		Name:     "localhost:27017",
		Health:   status.MemberHealthUp,
		State:    status.MemberStateRecovering,
		StateStr: "RECOVERING",
		Self:     true,
	}
	testStatus = &status.Status{
		Set:     "test",
		MyState: status.MemberStatePrimary,
		Ok:      1,
		Members: []*status.Member{
			testMember,
		},
	}
)

func TestGetSelfMemberState(t *testing.T) {
	state := getSelfMemberState(testStatus)
	if *state != testMember.State {
		t.Errorf("healthcheck.getSelfMemberState() returned %s instead of %s", *state, testMember.State)
	}
}

func TestIsMemberStateOk(t *testing.T) {
	state := getSelfMemberState(testStatus)
	if !isStateOk(state) {
		t.Errorf("healthcheck.isStateOk(\"%s\") returned false", *state)
	}

	testStatusFail := testStatus
	testStatusFail.Members[0].State = status.MemberStateRemoved
	stateFail := getSelfMemberState(testStatusFail)
	if isStateOk(stateFail) {
		t.Errorf("healthcheck.isStateOk(\"%s\") returned true", *stateFail)
	}
}
