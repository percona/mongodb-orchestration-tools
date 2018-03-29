package replset

import (
	"sync"
	"time"

	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"gopkg.in/mgo.v2"
)

type Replset struct {
	sync.Mutex
	config      *config.Config
	Members     map[string]*Mongod
	Name        string
	LastUpdated time.Time
}

func New(config *config.Config, name string) *Replset {
	return &Replset{
		config:  config,
		Members: make(map[string]*Mongod),
		Name:    name,
	}
}

func (r *Replset) UpdateMember(mongod *Mongod) {
	r.Lock()
	defer r.Unlock()
	r.Members[mongod.Name()] = mongod
	r.LastUpdated = time.Now()
}

func (r *Replset) RemoveMember(mongod *Mongod) {
	r.Lock()
	defer r.Unlock()
	delete(r.Members, mongod.Name())
}

func (r *Replset) GetMember(name string) *Mongod {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.Members[name]; ok {
		return r.Members[name]
	}
	return nil
}

func (r *Replset) HasMember(name string) bool {
	return r.GetMember(name) != nil
}

func (r *Replset) GetMembers() map[string]*Mongod {
	r.Lock()
	defer r.Unlock()
	return r.Members
}

func (r *Replset) GetReplsetDialInfo() *mgo.DialInfo {
	di := &mgo.DialInfo{
		Direct:         false,
		FailFast:       true,
		ReplicaSetName: r.Name,
		Timeout:        r.config.ReplsetTimeout,
	}
	for _, member := range r.GetMembers() {
		di.Addrs = append(di.Addrs, member.Name())
	}
	if r.config.Username != "" && r.config.Password != "" {
		di.Username = r.config.Username
		di.Password = r.config.Password
	}
	return di
}
