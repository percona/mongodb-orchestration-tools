package fetcher

import (
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
)

func TestWatchdogReplsetFetcherNew(t *gotesting.T) {
	testing.DoSkipTest(t)

	rsConfigManager := rsConfig.New(testDBSession)
	testFetcher = New(testDBSession, rsConfigManager)
	assert.NotNil(t, testFetcher, ".New() should not return nil")
}

func TestWatchdogReplsetFetcherGetConfig(t *gotesting.T) {
	testing.DoSkipTest(t)

	config, err := testFetcher.GetConfig()
	assert.NoError(t, err, ".GetConfig() should not return an error")
	assert.NotNil(t, config, ".GetConfig() should not return a nil config")

	assert.NotEmpty(t, config.Members, ".GetConfig() should not return a config with no members")
	assert.Equal(t, testing.MongodbReplsetName, config.Name, ".GetConfig() should not return a config with wrong name")
}

func TestWatchdogReplsetFetcherGetStatus(t *gotesting.T) {
	testing.DoSkipTest(t)

	status, err := testFetcher.GetStatus()
	assert.NoError(t, err, ".GetStatus() should not return an error")
	assert.NotNil(t, status, ".GetStatus() should not return a nil status")

	assert.NotEmpty(t, status.Members, ".GetStatus() should not return a status with no members")
	assert.Equal(t, testing.MongodbReplsetName, status.Set, ".GetStatus() should not return a status with wrong name")
}
