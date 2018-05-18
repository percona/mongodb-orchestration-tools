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

package common

import (
	"fmt"
	"runtime"

	"github.com/alecthomas/kingpin"
	dcosmongotools "github.com/percona/dcos-mongo-tools"
)

const appAuthor = "Percona LLC."

// NewApp sets up a kingpin.Application
func NewApp(app *kingpin.Application, commit, branch string) {
	app.Version(fmt.Sprintf(
		"%s version %s\ngit commit %s, branch %s\ngo version %s",
		app.Name, dcosmongotools.Version, commit, branch, runtime.Version(),
	))
	app.Author(appAuthor)
}

// DoStop checks if a goroutine should stop, based on a boolean channel
func DoStop(stop *chan bool) bool {
	select {
	case doStop := <-*stop:
		return doStop
	default:
		return false
	}
}
