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
