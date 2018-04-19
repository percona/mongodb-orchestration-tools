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

package user

import (
	"os"

	"github.com/percona/dcos-mongo-tools/common"
	"gopkg.in/mgo.v2"
)

var (
	ClusterAdminUsername   = os.Getenv(common.EnvMongoDBClusterAdminUser)
	ClusterAdminPassword   = os.Getenv(common.EnvMongoDBClusterAdminPassword)
	ClusterMonitorUsername = os.Getenv(common.EnvMongoDBClusterMonitorUser)
	ClusterMonitorPassword = os.Getenv(common.EnvMongoDBClusterMonitorPassword)
	BackupUsername         = os.Getenv(common.EnvMongoDBBackupUser)
	BackupPassword         = os.Getenv(common.EnvMongoDBBackupPassword)
	UserAdminUsername      = os.Getenv(common.EnvMongoDBUserAdminUser)
	UserAdminPassword      = os.Getenv(common.EnvMongoDBUserAdminPassword)

	SystemUserDatabase = "admin"
	UserAdmin          = &mgo.User{
		Username: UserAdminUsername,
		Password: UserAdminPassword,
		Roles: []mgo.Role{
			RoleUserAdminAny,
		},
	}
	SystemUsers = []*mgo.User{
		&mgo.User{
			Username: ClusterAdminUsername,
			Password: ClusterAdminPassword,
			Roles: []mgo.Role{
				RoleClusterAdmin,
			},
		},
		&mgo.User{
			Username: ClusterMonitorUsername,
			Password: ClusterMonitorPassword,
			Roles: []mgo.Role{
				RoleClusterMonitor,
			},
		},
		&mgo.User{
			Username: BackupUsername,
			Password: BackupPassword,
			Roles: []mgo.Role{
				RoleBackup,
				RoleClusterMonitor,
			},
		},
	}
	SystemUsernames = []string{
		UserAdminUsername,
		ClusterAdminUsername,
		ClusterMonitorUsername,
		BackupUsername,
	}
)
