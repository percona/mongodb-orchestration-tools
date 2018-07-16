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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DoStop checks if a goroutine should stop, based on a boolean channel
func DoStop(stop *chan bool) bool {
	select {
	case doStop := <-*stop:
		return doStop
	default:
		return false
	}
}

// GetUserID returns the numeric ID of a system user
func GetUserID(userName string) (int, error) {
	u, err := user.Lookup(userName)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(u.Uid)
}

// GetGroupID returns the numeric ID of a system group
func GetGroupID(groupName string) (int, error) {
	g, err := user.LookupGroup(groupName)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(g.Gid)
}

// RelPathToAbs returns a string containing a absolute to the provided path, relative to the caller directory
func RelPathToAbs(relPath string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	baseDir := filepath.Dir(filename)
	path, err := filepath.Abs(filepath.Join(baseDir, relPath))
	if err == nil {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// StringFromFile returns a string using the contents of an existing filename
func StringFromFile(fileName string) *string {
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err == nil {
		data := strings.TrimSpace(string(bytes))
		return &data
	}
	return nil
}

// PasswordFallbackFromFile loads a password from file if it exists
func PasswordFallbackFromFile(password string, passwordName string) string {
	if _, err := os.Stat(password); err == nil {
		log.Infof("Loading %s password from file %s", passwordName, password)
		str := StringFromFile(password)
		if str != nil {
			return *str
		}
	}
	return password
}
