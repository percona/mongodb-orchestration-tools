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

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/db"
)

var (
	DefaultAPIPoll               = "10s"
	DefaultDelayWatcher          = "20s"
	DefaultReplsetPoll           = "5s"
	DefaultReplsetTimeout        = "3s"
	DefaultReplsetConfUpdatePoll = "10s"
)

// Watchdog Configuration
type Config struct {
	Tool                  *common.ToolConfig
	Username              string
	Password              string
	FrameworkName         string
	API                   *api.Config
	APIPoll               time.Duration
	SSL                   *db.SSLConfig
	ReplsetPoll           time.Duration
	ReplsetTimeout        time.Duration
	ReplsetConfUpdatePoll time.Duration
	DelayWatcher          time.Duration
}
