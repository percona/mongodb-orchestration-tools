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
	"os"
	"path/filepath"
	"runtime"
)

// DoStop checks if a goroutine should stop, based on a boolean channel
func DoStop(stop *chan bool) bool {
	select {
	case doStop := <-*stop:
		return doStop
	default:
		return false
	}
}

// RelPath returns a string containing a absolute to the provided path, relative to the caller directory
func RelPathToAbs(relPath string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	baseDir := filepath.Dir(filename)
	path, err := filepath.Abs(filepath.Join(baseDir, relPath))
	if err == nil {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
