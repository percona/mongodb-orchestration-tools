package api

type Endpoints []string

type Endpoint struct {
	Address []string `json:"address"`
	Dns     []string `json:"dns"`
}

func (e *Endpoint) Addresses() []string {
	return e.Address
}

func (e *Endpoint) Hosts() []string {
	return e.Dns
}
