package job

import (
	"context"
	"fmt"
	"log"

	yaml "gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	w_tools "k8s.io/client-go/tools/watch"

	"time"
)

// waits for the job to complete
func WaitForJobToSucceeded(client kubernetes.Interface, namespace, jobName string, timeout time.Duration) error {
	job, err := client.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	w, err := client.BatchV1().Jobs(namespace).Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", job.Name).String(),
	})
	if err != nil {
		return err
	}

	defer w.Stop()
	condition := func(event watch.Event) (bool, error) {
		job, ok := event.Object.(*batchv1.Job)
		if !ok {
			fmt.Println("unexpected type")
		}
		return job.Status.Succeeded == 1, nil
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err = w_tools.UntilWithoutRetry(ctx, w, condition)
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("job %s never succeeded", jobName)
	}
	return nil
}

func WaitForJobToComplete(client kubernetes.Interface, namespace, jobName string, timeout time.Duration, verbose bool) error {
	job, err := client.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	options := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", job.Name).String()}

	w, err := client.BatchV1().Jobs(namespace).Watch(options)
	if err != nil {
		return err
	}

	defer w.Stop()

	condition := func(event watch.Event) (bool, error) {
		job := event.Object.(*batchv1.Job)
		completionTime := job.Status.CompletionTime
		complete := completionTime != nil && !completionTime.IsZero()
		if complete && verbose {
			data, _ := yaml.Marshal(job)
			log.Logger().Infof("Job %s is complete: %s", jobName, string(data))
		}
		return complete, nil
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err = tools_watch.UntilWithoutRetry(ctx, w, condition)
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("job %s never terminated", jobName)
	}
	return nil
}

func IsJobSucceeded(job *batchv1.Job) bool {
	return IsJobFinished(job) && job.Status.Succeeded > 0
}

func IsJobFinished(job *batchv1.Job) bool {
	BackoffLimit := job.Spec.BackoffLimit
	return job.Status.CompletionTime != nil || (job.Status.Active == 0 && BackoffLimit != nil && job.Status.Failed >= *BackoffLimit)
}

func DeleteJob(client kubernetes.Interface, namespace, name string) error {
	err := client.BatchV1().Jobs(namespace).Delete(name, metav1.NewDeleteOptions(0))
	if err != nil {
		return fmt.Errorf("error deleting job %s. error: %v", name, err)
	}
	return nil
}
