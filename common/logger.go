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
	"io"

	lcf "github.com/Robpol86/logrus-custom-formatter"
	log "github.com/sirupsen/logrus"
)

// GetLogFormatter returns a configured logrus.Formatter for logging
func GetLogFormatter(progName string) log.Formatter {
	template := "%[ascTime]s %-5[process]d " + progName + "  %-7[levelName]s %[message]s %[fields]s\n"
	return lcf.NewFormatter(template, nil)
}

// SetupLogger configures github.com/srupsen/logrus for logging
func SetupLogger(config *ToolConfig, formatter log.Formatter, out io.Writer) {
	log.SetOutput(out)
	log.SetFormatter(formatter)
	log.SetLevel(log.InfoLevel)
	if config.Verbose {
		log.SetLevel(log.DebugLevel)
	}
}
