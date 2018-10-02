package k8s

import "percona/mongodb-orchestration-tools/pkg/pod"

type Pod struct{}

func Name() string {
	return "k8s"
}

func GetPodURL() string {
	return "operator-sdk"
}

func GetPods() (*pod.Pods, error) {
	return &pod.Pods{}, nil
}

func GetPodTasks(podName string) ([]Task, error) {
	return []Task{}, nil
}
