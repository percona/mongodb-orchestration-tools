package common

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

func NewCommand(bin string, args []string, user, group string) (*Command, error) {
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
