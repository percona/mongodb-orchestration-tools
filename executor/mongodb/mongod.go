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

package mongodb

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
	mongo_config "github.com/timvaillancourt/go-mongodb-config/config"
)

const (
	DefaultDirMode = os.FileMode(0700)
	DefaultKeyMode = os.FileMode(0400)
)

type Mongod struct {
	config     *Config
	configFile string
	commandBin string
	command    *command.Command
}

func NewMongod(config *Config, nodeType string) *Mongod {
	return &Mongod{
		config:     config,
		configFile: filepath.Join(config.ConfigDir, nodeType+".conf"),
		commandBin: filepath.Join(config.BinDir, nodeType),
	}
}

func mkdir(path string, uid int, gid int, mode os.FileMode) error {
	if _, err := os.Stat(path); err != nil {
		err = os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	err := os.Chown(path, uid, gid)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongod) Initiate() error {
	uid, err := command.GetUserId(m.config.User)
	if err != nil {
		log.Errorf("Could not get user %s UID: %s\n", m.config.User, err)
		return err
	}

	gid, err := command.GetGroupId(m.config.Group)
	if err != nil {
		log.Errorf("Could not get group %s GID: %s\n", m.config.Group, err)
		return err
	}

	log.WithFields(log.Fields{
		"config": m.configFile,
	}).Info("Loading mongodb config file")
	config, err := mongo_config.Load(m.configFile)
	if err != nil {
		log.Errorf("Error loading mongodb configuration: %s", err)
		return err
	}
	if config.Security == nil || config.Security.KeyFile == "" || config.Storage == nil || config.Storage.DbPath == "" {
		return errors.New("mongodb config file is not valid, must have security.keyFile and storage.dbPath defined!")
	}

	log.WithFields(log.Fields{
		"tmpDir": m.config.TmpDir,
	}).Info("Initiating the mongod tmp dir")
	err = mkdir(m.config.TmpDir, uid, gid, DefaultDirMode)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"keyFile": config.Security.KeyFile,
	}).Info("Initiating the mongod keyFile")
	err = os.Chown(config.Security.KeyFile, uid, gid)
	if err != nil {
		return err
	}
	err = os.Chmod(config.Security.KeyFile, DefaultKeyMode)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"dbPath": config.Storage.DbPath,
	}).Info("Initiating the mongod dbPath")
	err = mkdir(config.Storage.DbPath, uid, gid, DefaultDirMode)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongod) IsStarted() bool {
	if m.command != nil {
		return m.command.IsRunning()
	}
	return false
}

func (m *Mongod) Start() error {
	err := m.Initiate()
	if err != nil {
		log.Errorf("Error initiating mongod environment on this host: %s", err)
		return err
	}

	m.command, err = command.New(
		m.commandBin,
		[]string{"--config", m.configFile},
		m.config.User,
		m.config.Group,
	)
	if err != nil {
		return err
	}

	return m.command.Start()
}

func (m *Mongod) Wait() {
	if m.command != nil && m.command.IsRunning() {
		m.command.Wait()
	}
}
