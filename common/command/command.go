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
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Command struct {
	Bin   string
	Args  []string
	Group string
	User  string

	command *exec.Cmd
	uid     int
	gid     int
	running bool
}

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

func New(bin string, args []string, user, group string) (*Command, error) {
	c := &Command{
		Bin:   bin,
		Args:  args,
		User:  user,
		Group: group,
	}
	return c, c.prepare()
}

func (c *Command) IsRunning() bool {
	return c.running
}

func (c *Command) prepare() error {
	var err error

	c.uid, err = GetUserId(c.User)
	if err != nil {
		return err
	}

	c.gid, err = GetGroupId(c.Group)
	if err != nil {
		return err
	}

	c.command = exec.Command(c.Bin, c.Args...)
	c.command.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(c.uid),
			Gid: uint32(c.gid),
		},
	}

	return nil
}

func (c *Command) Start() error {
	log.WithFields(log.Fields{
		"command": c.Bin,
		"args":    c.Args,
		"user":    c.User,
		"group":   c.Group,
	}).Debug("Starting command")

	c.command.Stdout = os.Stdout
	c.command.Stderr = os.Stderr

	err := c.command.Start()
	if err != nil {
		return err
	}
	c.running = true

	return nil
}

func (c *Command) CombinedOutput() ([]byte, error) {
	log.WithFields(log.Fields{
		"command": c.Bin,
		"args":    c.Args,
		"user":    c.User,
		"group":   c.Group,
	}).Debug("Running command")

	return c.command.CombinedOutput()
}

func (c *Command) Run() error {
	log.WithFields(log.Fields{
		"command": c.Bin,
		"args":    c.Args,
		"user":    c.User,
		"group":   c.Group,
	}).Debug("Running command")

	return c.command.Run()
}

func (c *Command) Wait() {
	if c.IsRunning() {
		c.command.Wait()
		c.running = false
	}
}

func (c *Command) Kill() error {
	if c.command.Process == nil {
		return nil
	}
	return c.command.Process.Kill()
}
