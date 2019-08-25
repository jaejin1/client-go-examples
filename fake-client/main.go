package main

import (
	"context"
	"log"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

var TestID = "test"
var TestNamespace = "test"

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kubernetesInterface := fake.NewSimpleClientset()

	objs := make(chan metav1.Object, 2)
	factory := informers.NewSharedInformerFactory(kubernetesInterface, 0)
	jobInformer := factory.Batch().V1().Jobs().Informer()
	podInformer := factory.Core().V1().Pods().Informer()

	jobInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			Obj := obj.(metav1.Object)
			log.Printf("job added: %s/%s", Obj.GetNamespace(), Obj.GetName())
			objs <- Obj
		},
	})

	podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			Obj := obj.(metav1.Object)
			log.Printf("pod added: %s/%s", Obj.GetNamespace(), Obj.GetName())
			objs <- Obj
		},
	})

	factory.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), jobInformer.HasSynced, podInformer.HasSynced)

	// fake pod data
	p := &apiv1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}}
	_, err := kubernetesInterface.CoreV1().Pods(TestNamespace).Create(p)
	if err != nil {
		log.Printf("error injecting pod add: %v", err)
	}

	// fake job data
	j := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestID,
			Namespace: TestNamespace,
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Image: "foo/bar"},
					},
				},
			},
		},
	}
	j.Status.Succeeded = 1
	_, err = kubernetesInterface.BatchV1().Jobs(TestNamespace).Create(j)
	if err != nil {
		log.Printf("error injecting job add: %v", err)
	}

	count := 0
	for {
		select {
		case obj := <-objs:
			log.Printf("Got pod from channel: %s/%s", obj.GetNamespace(), obj.GetName())
			count += 1
		case <-time.After(wait.ForeverTestTimeout):
			log.Printf("Informer did not get the added pod")
		default:
			if count == 2 {
				return
			}
		}
	}
}
