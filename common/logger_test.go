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

package common

import (
	"bytes"
	"os"
	"strings"
	gotesting "testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var toolConfig = &ToolConfig{
	ProgName: "test",
	Verbose:  true,
}

func TestSetupLogger(t *gotesting.T) {
	assert.Equal(t, log.InfoLevel, log.GetLevel(), "logrus.GetLevel() should return info level")
	formatter := GetLogFormatter(toolConfig.ProgName)
	SetupLogger(toolConfig, formatter, os.Stdout)
	assert.Equal(t, log.DebugLevel, log.GetLevel(), "logrus.GetLevel() should return debug level")
	assert.Equal(t, formatter, formatter, "logrus.StandarLogger().Formatter is incorrect")
}

func TestLogInfo(t *gotesting.T) {
	buf := new(bytes.Buffer)
	formatter := GetLogFormatter(toolConfig.ProgName)
	SetupLogger(toolConfig, formatter, buf)
	log.Info("test123")

	infoStr := strings.ToUpper(log.InfoLevel.String())
	expected := " " + toolConfig.ProgName + "  " + infoStr + "    test123 \n"
	logged := buf.String()
	assert.Truef(t,
		strings.HasSuffix(logged, expected),
		"log not equal got '%v' and expected '%v'",
		strings.TrimSpace(logged),
		strings.TrimSpace(expected),
	)
}
