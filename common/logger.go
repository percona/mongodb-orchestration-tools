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
	"os"

	lcf "github.com/Robpol86/logrus-custom-formatter"
	log "github.com/sirupsen/logrus"
)

func SetupLogger(config *ToolConfig) {
	template := "%[ascTime]s %-5[process]d " + config.ProgName + "  %-7[levelName]s %[message]s %[fields]s\n"
	log.SetOutput(os.Stdout)
	log.SetFormatter(lcf.NewFormatter(template, nil))
	log.SetLevel(log.InfoLevel)
	if config.Verbose {
		log.SetLevel(log.DebugLevel)
	}
}
