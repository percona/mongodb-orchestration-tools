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

package replset

import (
	"github.com/stretchr/testify/assert"
	gotesting "testing"
)

var (
	testManagerAddMongod = &Mongod{
		Host:    "testhost",
		Port:    123456,
		Replset: testReplsetName,
	}
)

func TestReplsetNewManager(t *gotesting.T) {
	testManager = NewManager(testWatchdogConfig)
	assert.NotNil(t, testManager, ".NewManager() should not return nil")
	assert.Len(t, testManager.replsets, 0, ".NewManager() should return a Manager with an empty slice of replsets")
}

func TestReplsetManagerGetFalse(t *gotesting.T) {
	assert.Nil(t, testManager.Get(testReplsetName), ".Get() should return nil")
}

func TestReplsetManagerHasReplsetFalse(t *gotesting.T) {
	assert.False(t, testManager.HasReplset(testReplsetName), ".HasReplset() should return false")
}

func TestReplsetManagerAddReplset(t *gotesting.T) {
	replset := New(testWatchdogConfig, testReplsetName)
	testManager.addReplset(replset)
	assert.True(t, testManager.HasReplset(testReplsetName), ".HasReplset() after .addReplset() should return true")
}

func TestReplsetManagerGetAll(t *gotesting.T) {
	assert.Len(t, testManager.GetAll(), 1, ".GetAll() should return a single Replset")
}

func TestReplsetManagerGet(t *gotesting.T) {
	assert.NotNil(t, testManager.Get(testReplsetName), ".Get() should return a Replset")
}

func TestReplsetManagerUpdateMember(t *gotesting.T) {
	testManager.UpdateMember(&Mongod{
		Host:    "anotherhost",
		Port:    12241,
		Replset: "anotherRs",
	})

	testManager.UpdateMember(testManagerAddMongod)
	rs := testManager.Get(testReplsetName)
	assert.NotNil(t, rs, ".Get() should not return nil")
	assert.True(t, rs.HasMember(testManagerAddMongod.Name()), ".HasMember() on replset returned false after .UpdateMember()")
}

func TestReplsetManagerHasMember(t *gotesting.T) {
	assert.True(t, testManager.HasMember(testManagerAddMongod), ".HasMember() on replset returned false after .UpdateMember()")
	assert.False(t, testManager.HasMember(&Mongod{Host: "doesntexit", Port: 12345, Replset: "notexists"}), ".HasMember() for missing member should return false")
}

func TestReplsetManagerRemoveMember(t *gotesting.T) {
	testManager.RemoveMember(testManagerAddMongod)
	assert.False(t, testManager.HasMember(testManagerAddMongod), ".HasMember() on replset returned true after .RemoveMember()")
}
