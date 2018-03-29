package common

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
)

type ToolConfig struct {
	ProgName     string
	Verbose      bool
	PrintVersion bool
}

func NewToolConfig(progName string) *ToolConfig {
	tool := &ToolConfig{
		ProgName: filepath.Base(progName),
	}
	kingpin.Flag("verbose", "log verbose information").BoolVar(&tool.Verbose)
	kingpin.Flag("version", "print version info and exit").BoolVar(&tool.PrintVersion)
	return tool
}

func (tc *ToolConfig) PrintVersionAndExit() {
	PrintVersion(tc.ProgName)
	os.Exit(0)
}
