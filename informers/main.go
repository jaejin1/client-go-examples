package main

import (
	"flag"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	factory := informers.NewFilteredSharedInformerFactory(clientset, 0, "default", func(opt *metav1.ListOptions) {
		opt.LabelSelector = "app"
		//opt.LabelSelector = "app=applabel"
	})
	// if you want to get all pods in namespace
	//factory := informers.NewFilteredSharedInformerFactory(clientset, 0, "default", func(opt *v1.ListOptions) {})
	// if you want to get all namespace
	//factory := informers.NewSharedInformerFactory(clientset, 0)

	informer := factory.Core().V1().Pods().Informer()
	//informer := factory.Core().V1().Services().Informer()
	//informer := factory.Batch().V1().Jobs().Informer()
	//informer := factory.Apps().V1().Deployments().Informer()
	//...

	stopper := make(chan struct{})
	defer close(stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			Obj := obj.(metav1.Object)
			log.Printf("New Pod Added to Store: %s in %s", Obj.GetName(), Obj.GetNamespace())
		},
	})

	informer.Run(stopper)
}
