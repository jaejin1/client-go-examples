package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	tools_watch "k8s.io/client-go/tools/watch"
)

func GetPods(client kubernetes.Interface, ns string, filter string) ([]string, map[string]*v1.Pod, error) {
	names := []string{}
	m := map[string]*v1.Pod{}
	list, err := client.CoreV1().Pods(ns).List(meta_v1.ListOptions{})
	if err != nil {
		return names, m, fmt.Errorf("Failed to load Pods %s", err)
	}
	for _, d := range list.Items {
		c := d
		name := d.Name
		m[name] = &c
		if filter == "" || strings.Contains(name, filter) && d.DeletionTimestamp == nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, m, nil
}

func IsPodReady(pod *v1.Pod) bool {
	phase := pod.Status.Phase
	if phase != v1.PodRunning || pod.DeletionTimestamp != nil {
		return false
	}
	return IsPodReadyConditionTrue(pod.Status)
}

func IsPodCompleted(pod *v1.Pod) bool {
	phase := pod.Status.Phase
	if phase == v1.PodSucceeded || phase == v1.PodFailed {
		return true
	}
	return false
}

func IsPodReadyConditionTrue(status v1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == v1.ConditionTrue
}

func PodStatus(pod *v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	phase := pod.Status.Phase
	if IsPodReady(pod) {
		return "Ready"
	}
	return string(phase)
}

func GetPodReadyCondition(status v1.PodStatus) *v1.PodCondition {
	_, condition := GetPodCondition(&status, v1.PodReady)
	return condition
}

func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

func waitForPodSelector(client kubernetes.Interface, namespace string, options meta_v1.ListOptions,
	timeout time.Duration, condition func(event watch.Event) (bool, error)) error {
	w, err := client.CoreV1().Pods(namespace).Watch(options)
	if err != nil {
		return err
	}
	defer w.Stop()

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err = tools_watch.UntilWithoutRetry(ctx, w, condition)

	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("pod %s never became ready", options.String())
	}
	return nil
}

/*
func HasContainerStarted(pod *v1.Pod, idx int) bool {
	if pod == nil {
		return false
	}
	_, statuses, _ := GetContainersWithStatusAndIsInit(pod)
	if idx >= len(statuses) {
		return false
	}
	ic := statuses[idx]
	if ic.State.Running != nil || ic.State.Terminated != nil {
		return true
	}
	return false
}
*/

func GetPodNames(client kubernetes.Interface, ns string, filter string) ([]string, error) {
	names := []string{}
	list, err := client.CoreV1().Pods(ns).List(meta_v1.ListOptions{})
	if err != nil {
		return names, fmt.Errorf("Failed to load Pods %s", err)
	}
	for _, d := range list.Items {
		name := d.Name
		if filter == "" || strings.Contains(name, filter) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func GetPodRestarts(pod *v1.Pod) int32 {
	var restarts int32
	statuses := pod.Status.ContainerStatuses
	if len(statuses) == 0 {
		return restarts
	}
	for _, status := range statuses {
		restarts += status.RestartCount
	}
	return restarts
}
