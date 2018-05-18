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
	"os"
	"path/filepath"
	"runtime"
	gotesting "testing"
	"time"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
)

const testSSLDirRelPath = "../../test/ssl"

func findTestSSLDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	baseDir := filepath.Dir(filename)
	path, err := filepath.Abs(filepath.Join(baseDir, testSSLDirRelPath))
	if err == nil {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

var (
	sslCertPath = filepath.Join(findTestSSLDir(), "client.pem")
	sslCAPath   = filepath.Join(findTestSSLDir(), "rootCA.crt")
)

func TestGetSessionSSL(t *gotesting.T) {
	testPrimaryDbConfigSSL := &Config{
		DialInfo: testPrimaryDbConfig.DialInfo,
		SSL: &SSLConfig{
			Enabled:    true,
			PEMKeyFile: sslCertPath,
			CAFile:     sslCAPath,
			Insecure:   false,
		},
	}

	// intentionally test for SSL error (due to self-signed SSL certs) in secure mode
	testLogBuffer.Reset()
	assert.Nil(t, LastSSLError(), ".LastSSLError() should be nil")
	testPrimaryDbConfigSSL.DialInfo.Timeout = 100 * time.Millisecond
	_, err := GetSession(testPrimaryDbConfigSSL)
	assert.Error(t, err, ".GetSession() should return an error due to self-signed certificates")
	assert.Error(t, LastSSLError(), ".LastSSLError() should not be nil")
	assert.Regexp(t, "^x509: cannot validate certificate for", LastSSLError().Error(), ".LastSSLError() has unexpected error message")
	assert.Contains(t, testLogBuffer.String(), "x509: cannot validate certificate for", ".GetSession() log output should contain ssl error")

	// enable insecure mode (due to self-signed certs) and connect
	testPrimaryDbConfigSSL.DialInfo.Timeout = testing.MongodbTimeout
	testPrimaryDbConfigSSL.SSL.Insecure = true
	testPrimarySessionSSL, err := GetSession(testPrimaryDbConfigSSL)
	assert.NoError(t, err, ".GetSession() should return no error")
	defer testPrimarySessionSSL.Close()

	// test SSL connection
	assert.NotNil(t, testPrimarySessionSSL, ".GetSession() should not return a nil testPrimarySession")
	assert.NoError(t, testPrimarySessionSSL.Ping(), ".GetSession() returned a session that failed to ping")
}
