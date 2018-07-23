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
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"

	"gopkg.in/mgo.v2"
)

const (
	adminDB = "admin"
)

type UserRoleJSON struct {
	Role     string `json:"role"`
	Database string `json:"db"`
}

type UserJSON struct {
	Username string          `json:"user"`
	Password string          `json:"pwd"`
	Roles    []*UserRoleJSON `json:"roles"`
}

type UserChangeJSON struct {
	Users []*UserJSON `json:"users"`
}

func (user *UserJSON) Validate(db string) error {
	if user.Username == "" {
		return errors.New("'user' field is required")
	} else if user.Password == "" {
		return errors.New("'pwd' field is required")
	} else if len(user.Roles) > 0 {
		for _, role := range user.Roles {
			if role.Role == "" {
				return errors.New("'role' field is required")
			} else if role.Database == "" {
				return errors.New("'db' field is required")
			} else if role.Database != db && db != adminDB {
				return errors.New("cannot set role for other database unless user is added to 'admin'")
			}
		}
	} else {
		return errors.New("'roles' field is required. must be an array with one or more role documents!")
	}
	return nil
}

func NewFromJSONFile(file string) (*UserJSON, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	user := &UserJSON{}
	err = json.Unmarshal(bytes, user)
	return user, err
}

func NewFromCLIPayloadFile(file string) ([]*UserJSON, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(bytes)))
	_, err = base64.StdEncoding.Decode(decoded, bytes)
	if err != nil {
		return nil, err
	}

	user := &UserChangeJSON{}
	err = json.Unmarshal(decoded, user)
	return user.Users, err
}

func (user *UserJSON) ToMgoUser(db string) (*mgo.User, error) {
	err := user.Validate(db)
	if err != nil {
		return nil, err
	}

	roles := []mgo.Role{}
	otherDBRoles := map[string][]mgo.Role{}
	for _, role := range user.Roles {
		if role.Database == db {
			roles = append(roles, mgo.Role(role.Role))
			continue
		} else if db == adminDB {
			otherDBRoles[role.Database] = append(otherDBRoles[role.Database], mgo.Role(role.Role))
		}
	}
	return &mgo.User{
		Username:     user.Username,
		Password:     user.Password,
		Roles:        roles,
		OtherDBRoles: otherDBRoles,
	}, nil
}
