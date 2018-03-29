package user

import (
	"encoding/base64"
	"errors"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	RoleBackup         mgo.Role = "backup"
	RoleClusterAdmin   mgo.Role = mgo.RoleClusterAdmin
	RoleClusterMonitor mgo.Role = "clusterMonitor"
	RoleUserAdminAny   mgo.Role = mgo.RoleUserAdminAny
)

type UserChangeData struct {
	Users []*mgo.User `bson:users"`
}

func UpdateUser(session *mgo.Session, user *mgo.User, dbName string) error {
	if user.Username == "" || user.Password == "" {
		return errors.New("No username or password for user admin!")
	}
	if len(user.Roles) <= 0 {
		return errors.New("No roles defined for user!")
	}

	log.WithFields(log.Fields{
		"user":  user.Username,
		"roles": user.Roles,
		"db":    dbName,
	}).Info("Adding/updating MongoDB user")

	return session.DB(dbName).UpsertUser(user)
}

func UpdateUsers(session *mgo.Session, users []*mgo.User, dbName string) error {
	for _, user := range users {
		err := UpdateUser(session, user, dbName)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveUser(session *mgo.Session, username, db string) error {
	log.Infof("Removing user %s from db %s", username, db)
	err := session.DB(db).RemoveUser(username)
	if err == mgo.ErrNotFound {
		log.Warnf("Cannot remove user, %s does not exist in database %s", username, db)
		return nil
	}
	return err
}

func loadFromBase64BSONFile(file string) (*UserChangeData, error) {
	payload := &UserChangeData{}

	log.Infof("Loading mongodb user(s) from file %s", file)

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return payload, err
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(bytes)))
	_, err = base64.StdEncoding.Decode(decoded, bytes)
	if err != nil {
		return payload, err
	}

	err = bson.Unmarshal(decoded, payload)
	return payload, err
}

func IsSystemUser(username, db string) bool {
	if db != SystemUserDatabase {
		return false
	}
	for _, sysUsername := range SystemUsernames {
		if username == sysUsername {
			return true
		}
	}
	return false
}
