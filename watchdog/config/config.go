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

package config

import (
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/dcos/api"
)

var (
	DefaultDCOSAPIPoll    = "10s"
	DefaultDCOSIgnorePods = []string{
		"admin-0",
		"restore-0",
		"mongodb-consistent-backup-0",
	}
	DefaultReplsetPoll    = "5s"
	DefaultReplsetTimeout = "3s"
)

type ConfigDCOS struct {
	API           *api.Config
	APIPoll       time.Duration
	FrameworkName string
}

// Watchdog Configuration
type Config struct {
	Username       string
	Password       string
	IgnorePods     []string
	SSL            *db.SSLConfig
	ReplsetPoll    time.Duration
	ReplsetTimeout time.Duration
	SourcePoll     time.Duration
	MetricsPort    string
	DCOS           *ConfigDCOS
}
