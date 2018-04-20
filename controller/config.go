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

package controller

import (
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/db"
)

var (
	DefaultPodName              = "mongo"
	DefaultInitDelay            = "15s"
	DefaultRetrySleep           = "3s"
	DefaultMaxConnectTries      = "30"
	DefaultInitMaxReplTries     = "60"
	DefaultInitMaxAddUsersTries = "60"
)

type ConfigReplsetInit struct {
	PrimaryAddr      string
	MongoDBPort      string
	Delay            time.Duration
	MaxConnectTries  uint
	MaxReplTries     uint
	MaxAddUsersTries uint
	RetrySleep       time.Duration
}

type ConfigUser struct {
	API             *api.Config
	Endpoint        *api.Endpoint
	EndpointName    string
	Database        string
	Username        string
	File            string
	MaxConnectTries uint
	RetrySleep      time.Duration
}

type Config struct {
	DB                *db.Config
	Tool              *common.ToolConfig
	FrameworkName     string
	Replset           string
	UserAdminUser     string
	UserAdminPassword string
	ReplsetInit       *ConfigReplsetInit
	User              *ConfigUser
}
