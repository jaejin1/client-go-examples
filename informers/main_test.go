package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// TestFakeClient demonstrates how to use a fake client with SharedInformerFactory in tests.
func TestFakeClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the fake client.
	client := fake.NewSimpleClientset()

	// We will create an informer that writes added pods to a channel.
	objs := make(chan metav1.Object, 1)

	informers := informers.NewSharedInformerFactory(client, 0)
	podInformer := informers.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			Obj := obj.(metav1.Object)
			t.Logf("pod added: %s/%s", Obj.GetNamespace(), Obj.GetName())
			objs <- Obj
		},
	})

	// Make sure informers are running.
	informers.Start(ctx.Done())

	// This is not required in tests, but it serves as a proof-of-concept by
	// ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced)

	// Inject an event into the fake client.
	p := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}}
	_, err := client.CoreV1().Pods("test-ns").Create(p)
	fmt.Println("test")
	if err != nil {
		t.Fatalf("error injecting pod add: %v", err)
	}

	select {
	case obj := <-objs:
		t.Logf("Got pod from channel: %s/%s", obj.GetNamespace(), obj.GetName())
	case <-time.After(wait.ForeverTestTimeout):
		t.Error("Informer did not get the added pod")
	}
}