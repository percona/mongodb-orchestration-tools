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

package command

import (
	"os/user"
	"strconv"
)

// GetUserId returns the numeric ID of a system user
func GetUserId(userName string) (int, error) {
	u, err := user.Lookup(userName)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(u.Uid)
}

// GetGroupID returns the numeric ID of a system group
func GetGroupId(groupName string) (int, error) {
	g, err := user.LookupGroup(groupName)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(g.Gid)
}
