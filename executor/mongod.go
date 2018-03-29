package executor

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

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
	command    *exec.Cmd
	uid        int
	gid        int
	started    bool
}

func NewMongod(config *Config) *Mongod {
	return &Mongod{
		config:     config,
		configFile: filepath.Join(config.ConfigDir, config.NodeType+".conf"),
		commandBin: filepath.Join(config.BinDir, config.NodeType),
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

func (m *Mongod) getUidAndGid() (int, int, error) {
	u, err := user.Lookup(m.config.User)
	if err != nil {
		return -1, -1, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, err
	}

	g, err := user.LookupGroup(m.config.Group)
	if err != nil {
		return -1, -1, err
	}
	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return -1, -1, err
	}

	return uid, gid, nil
}

func (m *Mongod) Initiate() error {
	var err error

	m.uid, m.gid, err = m.getUidAndGid()
	if err != nil {
		log.Errorf("Error finding configured 'user' or 'group' on this host: %s", err)
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
	err = mkdir(m.config.TmpDir, m.uid, m.gid, DefaultDirMode)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"keyFile": config.Security.KeyFile,
	}).Info("Initiating the mongod keyFile")
	err = os.Chown(config.Security.KeyFile, m.uid, m.gid)
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
	err = mkdir(config.Storage.DbPath, m.uid, m.gid, DefaultDirMode)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongod) IsStarted() bool {
	return m.started
}

func (m *Mongod) Start() error {
	err := m.Initiate()
	if err != nil {
		log.Errorf("Error initiating mongod environment on this host: %s", err)
		return err
	}

	log.WithFields(log.Fields{
		"bin":    m.commandBin,
		"config": m.configFile,
		"group":  m.config.Group,
		"user":   m.config.User,
	}).Info("Starting mongod daemon")

	m.command = exec.Command(
		m.commandBin,
		"--config", m.configFile,
	)
	m.command.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(m.uid),
			Gid: uint32(m.gid),
		},
	}
	m.command.Stdout = os.Stdout
	m.command.Stderr = os.Stderr

	err = m.command.Start()
	if err != nil {
		return err
	}
	m.started = true

	return nil
}

func (m *Mongod) Wait() {
	if m.command != nil && m.IsStarted() {
		m.command.Wait()
	}
}
