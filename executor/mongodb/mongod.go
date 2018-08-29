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
	"io/ioutil"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
	mongoConfig "github.com/timvaillancourt/go-mongodb-config/config"
)

const (
	DefaultDirMode                   = os.FileMode(0700)
	DefaultKeyMode                   = os.FileMode(0400)
	minWiredTigerCacheSizeGB         = 0.25
	noMemoryLimit            int64   = 9223372036854771712
	gigaByte                 float64 = 1024 * 1024 * 1024
)

// getMemoryLimitBytes() returns the memory limit of a cgroup
// or zero/0 if there is no limit
func getMemoryLimitBytes() int64 {
	file := "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return 0
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return 0
	}
	limit32, _ := strconv.Atoi(strings.TrimSpace(string(bytes)))
	limit := int64(limit32)
	if limit == noMemoryLimit {
		return 0
	}
	return limit
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

type Mongod struct {
	config     *Config
	configFile string
	commandBin string
	command    *command.Command
}

func NewMongod(config *Config) *Mongod {
	return &Mongod{
		config:     config,
		configFile: filepath.Join(config.ConfigDir, "mongod.conf"),
		commandBin: filepath.Join(config.BinDir, "mongod"),
	}
}

// The WiredTiger internal cache, by default, will use the larger of either 50% of
// (RAM - 1 GB), or 256 MB. For example, on a system with a total of 4GB of RAM the
// WiredTiger cache will use 1.5GB of RAM (0.5 * (4 GB - 1 GB) = 1.5 GB).
//
// https://docs.mongodb.com/manual/reference/configuration-options/#storage.wiredTiger.engineConfig.cacheSizeGB
//
func (m *Mongod) getWiredTigerCacheSizeGB(limitBytes int64) float64 {
	size := math.Floor(m.config.WiredTigerCacheRatio * (float64(limitBytes) - gigaByte))
	sizeGB := size / gigaByte
	if sizeGB < minWiredTigerCacheSizeGB {
		sizeGB = minWiredTigerCacheSizeGB
	}
	return sizeGB
}

func (m *Mongod) Name() string {
	return "mongod"
}

func (m *Mongod) Initiate() error {
	uid, err := common.GetUserID(m.config.User)
	if err != nil {
		log.Errorf("Could not get user %s UID: %s\n", m.config.User, err)
		return err
	}

	gid, err := common.GetGroupID(m.config.Group)
	if err != nil {
		log.Errorf("Could not get group %s GID: %s\n", m.config.Group, err)
		return err
	}

	log.WithFields(log.Fields{
		"config": m.configFile,
	}).Info("Loading mongodb config file")
	config, err := mongoConfig.Load(m.configFile)
	if err != nil {
		log.Errorf("Error loading mongodb configuration: %s", err)
		return err
	}
	if config.Security == nil || config.Security.KeyFile == "" || config.Storage == nil || config.Storage.DbPath == "" {
		return errors.New("mongodb config file is not valid, must have security.keyFile and storage.dbPath defined!")
	}

	if config.Storage.Engine == "wiredTiger" {
		limitBytes := getMemoryLimitBytes()
		if limitBytes > 0 {
			cacheSizeGB := m.getWiredTigerCacheSizeGB(limitBytes)
			log.WithFields(log.Fields{
				"size_gb": cacheSizeGB,
				"ratio":   m.config.WiredTigerCacheRatio,
			}).Infof("Setting WiredTiger cache size")

			if config.Storage.WiredTiger == nil {
				config.Storage.WiredTiger = &mongoConfig.StorageWiredTiger{}
			}
			if config.Storage.WiredTiger.EngineConfig == nil {
				config.Storage.WiredTiger.EngineConfig = &mongoConfig.StorageWiredTigerEngineConfig{}
			}
			config.Storage.WiredTiger.EngineConfig.CacheSizeGB = cacheSizeGB

			err = config.Write(m.configFile)
			if err != nil {
				log.Errorf("Error writing new mongodb configuration: %s", err)
				return err
			}
		}
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

	mongodUser, err := user.Lookup(m.config.User)
	if err != nil {
		return err
	}

	mongodGroup, err := user.LookupGroup(m.config.Group)
	if err != nil {
		return err
	}

	m.command, err = command.New(
		m.commandBin,
		[]string{"--config", m.configFile},
		mongodUser,
		mongodGroup,
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

func (m *Mongod) Kill() error {
	if m.command == nil {
		return nil
	}
	return m.command.Kill()
}
