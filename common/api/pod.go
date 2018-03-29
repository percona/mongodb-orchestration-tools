package api

type Pods []string

type PodTask struct {
	Info   *PodTaskInfo   `json:"info"`
	Status *PodTaskStatus `json:"status"`
}

type PodTaskCommandEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PodTaskCommandEnvironment struct {
	Variables []*PodTaskCommandEnvironmentVariable `json:"variables"`
}

type PodTaskCommand struct {
	Environment *PodTaskCommandEnvironment `json:"environment"`
	Value       string                     `json:"value"`
}

type PodTaskInfo struct {
	Name    string          `json:"name"`
	Command *PodTaskCommand `json:"command"`
}

type PodTaskStatus struct {
	State *PodTaskState `json:"state"`
}
