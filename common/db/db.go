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

package db

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var (
	ErrMsgAuthFailedStr string = "server returned error on SASL authentication step: Authentication failed."
)

func GetSession(cnf *Config) (*mgo.Session, error) {
	log.WithFields(log.Fields{
		"hosts":      cnf.DialInfo.Addrs,
		"ssl":        cnf.SSL.Enabled,
		"ssl_secure": !cnf.SSL.Insecure,
	}).Debug("Connecting to mongodb")

	if cnf.DialInfo.Username != "" && cnf.DialInfo.Password != "" {
		log.WithFields(log.Fields{
			"user":   cnf.DialInfo.Username,
			"source": cnf.DialInfo.Source,
		}).Debug("Enabling authentication for session")
	}

	if cnf.SSL.Enabled {
		err := cnf.configureSSLDialInfo()
		if err != nil {
			log.Errorf("Failed to configure SSL/TLS: %s", err)
			return nil, err
		}
	}

	session, err := mgo.DialWithInfo(cnf.DialInfo)
	if err != nil && err.Error() == ErrMsgAuthFailedStr {
		log.Debug("Authentication failed, retrying with authentication disabled")
		cnf.DialInfo.Username = ""
		cnf.DialInfo.Password = ""
		session, err = mgo.DialWithInfo(cnf.DialInfo)
	}
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	return session, err
}

func WaitForSession(cnf *Config, maxRetries uint, sleepDuration time.Duration) (*mgo.Session, error) {
	var err error
	var tries uint
	for tries <= maxRetries || maxRetries == 0 {
		session, err := GetSession(cnf)
		if err == nil && session.Ping() == nil {
			return session, err
		}
		time.Sleep(sleepDuration)
		tries += 1
	}
	return nil, err
}
