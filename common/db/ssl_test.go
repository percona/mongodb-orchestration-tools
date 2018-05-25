// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"path/filepath"
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
)

const testSSLDirRelPath = "../../test/ssl"

var (
	sslCertFile = common.RelPathToAbs(filepath.Join(testSSLDirRelPath, "client.pem"))
	sslCAFile   = common.RelPathToAbs(filepath.Join(testSSLDirRelPath, "rootCA.crt"))
)

func TestDbLoadCaCertificate(t *gotesting.T) {
	sslConfig := &SSLConfig{
		Enabled: true,
		CAFile:  sslCAFile,
	}

	pool, err := sslConfig.loadCaCertificate()
	assert.NoError(t, err, ".loadCaCertificate() should return no error")
	assert.NotEmpty(t, pool.Subjects(), ".loadCaCertificate() should return a non-empty x509.CertPool")

	sslConfig.CAFile = "/does/not/exist.crt"
	_, err = sslConfig.loadCaCertificate()
	assert.Error(t, err, ".loadCaCertificate() should return an error when given missing path")
}

func TestDbConfigureSSLDialInfo(t *gotesting.T) {
	config := &Config{
		DialInfo: testPrimaryDbConfig.DialInfo,
		SSL: &SSLConfig{
			Enabled:    true,
			PEMKeyFile: sslCertFile,
			CAFile:     sslCAFile,
			Insecure:   true,
		},
	}
	assert.Nil(t, config.DialInfo.DialServer, "config.DialInfo.DialServer should be nil")

	err := config.configureSSLDialInfo()
	assert.NoError(t, err, ".configureSSLDialInfo() should not return an error")
	assert.NotNil(t, config.DialInfo.DialServer, "config.DialInfo.DialServer should not be nil")
}

func TestDbGetSessionSSL(t *gotesting.T) {
	testing.DoSkipTest(t)

	testPrimaryDbConfig.SSL = &SSLConfig{
		Enabled:    true,
		PEMKeyFile: sslCertFile,
		CAFile:     sslCAFile,
		Insecure:   false,
	}

	// intentionally test for SSL error (due to self-signed SSL certs) in secure mode
	testLogBuffer.Reset()
	assert.Nil(t, LastSSLError(), ".LastSSLError() should be nil")
	testPrimaryDbConfig.DialInfo.Timeout = 100 * time.Millisecond
	_, err := GetSession(testPrimaryDbConfig)
	assert.Error(t, err, ".GetSession() should return an error due to self-signed certificates")
	assert.Error(t, LastSSLError(), ".LastSSLError() should not be nil")
	assert.Regexp(t, "^x509: cannot validate certificate for", LastSSLError().Error(), ".LastSSLError() has unexpected error message")
	assert.Contains(t, testLogBuffer.String(), "x509: cannot validate certificate for", ".GetSession() log output should contain ssl error")

	// enable insecure mode (due to self-signed certs) and connect
	testPrimaryDbConfig.DialInfo.Timeout = testing.MongodbTimeout
	testPrimaryDbConfig.SSL.Insecure = true
	testPrimarySessionSSL, err := GetSession(testPrimaryDbConfig)
	assert.NoError(t, err, ".GetSession() should return no error")
	defer testPrimarySessionSSL.Close()

	// test SSL connection
	assert.NotNil(t, testPrimarySessionSSL, ".GetSession() should not return a nil testPrimarySession")
	assert.NoError(t, testPrimarySessionSSL.Ping(), ".GetSession() returned a session that failed to ping")

	testPrimaryDbConfig.SSL = &SSLConfig{}
}
