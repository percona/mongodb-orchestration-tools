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
	"path/filepath"

	"github.com/alecthomas/kingpin"
)

type ToolConfig struct {
	ProgName     string
	Verbose      bool
	PrintVersion bool
}

func NewToolConfig(progName string) *ToolConfig {
	tool := &ToolConfig{
		ProgName: filepath.Base(progName),
	}
	kingpin.Flag("verbose", "log verbose information").BoolVar(&tool.Verbose)
	kingpin.Flag("version", "print version info and exit").BoolVar(&tool.PrintVersion)
	return tool
}

func (tc *ToolConfig) PrintVersionAndExit() {
	PrintVersion(tc.ProgName)
	os.Exit(0)
}
