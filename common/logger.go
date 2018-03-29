package common

import (
	"os"

	lcf "github.com/Robpol86/logrus-custom-formatter"
	log "github.com/sirupsen/logrus"
)

func SetupLogger(config *ToolConfig) {
	template := "%[ascTime]s %-5[process]d " + config.ProgName + "  %-7[levelName]s %[message]s %[fields]s\n"
	log.SetOutput(os.Stdout)
	log.SetFormatter(lcf.NewFormatter(template, nil))
	log.SetLevel(log.InfoLevel)
	if config.Verbose {
		log.SetLevel(log.DebugLevel)
	}
}
