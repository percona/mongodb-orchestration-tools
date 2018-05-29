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

package fetcher

import (
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

type Fetcher interface {
	GetConfig() (*rsConfig.Config, error)
	GetStatus() (*rsStatus.Status, error)
}

type StateFetcher struct {
	configManager rsConfig.Manager
	session       *mgo.Session
}

func New(session *mgo.Session, configManager rsConfig.Manager) *StateFetcher {
	return &StateFetcher{
		configManager: configManager,
		session:       session,
	}
}

func (sf *StateFetcher) GetConfig() (*rsConfig.Config, error) {
	err := sf.configManager.Load()
	if err != nil {
		return nil, err
	}
	return sf.configManager.Get(), nil
}

func (sf *StateFetcher) GetStatus() (*rsStatus.Status, error) {
	return rsStatus.New(sf.session)
}
