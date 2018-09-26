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

	"github.com/percona/mongodb-orchestration-tools/internal"
	"gopkg.in/mgo.v2"
)

var (
	clusterAdminUsername   = os.Getenv(internal.EnvMongoDBClusterAdminUser)
	clusterAdminPassword   = os.Getenv(internal.EnvMongoDBClusterAdminPassword)
	clusterMonitorUsername = os.Getenv(internal.EnvMongoDBClusterMonitorUser)
	clusterMonitorPassword = os.Getenv(internal.EnvMongoDBClusterMonitorPassword)
	backupUsername         = os.Getenv(internal.EnvMongoDBBackupUser)
	backupPassword         = os.Getenv(internal.EnvMongoDBBackupPassword)
	userAdminUsername      = os.Getenv(internal.EnvMongoDBUserAdminUser)
	userAdminPassword      = os.Getenv(internal.EnvMongoDBUserAdminPassword)

	SystemUserDatabase = "admin"
	UserAdmin          = &mgo.User{
		Username: userAdminUsername,
		Password: userAdminPassword,
		Roles: []mgo.Role{
			RoleUserAdminAny,
		},
	}
	SystemUsers = []*mgo.User{
		{
			Username: clusterAdminUsername,
			Password: clusterAdminPassword,
			Roles: []mgo.Role{
				RoleClusterAdmin,
			},
		},
		{
			Username: clusterMonitorUsername,
			Password: clusterMonitorPassword,
			Roles: []mgo.Role{
				RoleClusterMonitor,
			},
		},
		{
			Username: backupUsername,
			Password: backupPassword,
			Roles: []mgo.Role{
				RoleBackup,
				RoleClusterMonitor,
				RoleRestore,
			},
		},
	}
	SystemUsernames = []string{
		userAdminUsername,
		clusterAdminUsername,
		clusterMonitorUsername,
		backupUsername,
	}
)
