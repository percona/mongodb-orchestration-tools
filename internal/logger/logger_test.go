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

package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInternalLoggerSetupLogger(t *testing.T) {
	assert.Equal(t, log.InfoLevel, log.GetLevel(), "logrus.GetLevel() should return info level")
	formatter := GetLogFormatter("test")
	SetupLogger(nil, formatter, os.Stdout)
	assert.Equal(t, formatter, formatter, "logrus.StandarLogger().Formatter is incorrect")
}

func TestInternalLoggerLogInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	formatter := GetLogFormatter("test")
	SetupLogger(nil, formatter, buf)
	log.Info("test123")

	infoStr := strings.ToUpper(log.InfoLevel.String())
	logged := buf.String()
	assert.Contains(t, strings.TrimSpace(logged), "test  logger_test.go:38 "+infoStr+"   test123", ".Info() log output unexpected")
}
