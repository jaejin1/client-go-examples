package statefulsets

import (
	"github.com/google/martian/log"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateStatefulSets(kubeClient kubernetes.Interface, sfs *appsv1.StatefulSet, namespace string) (*appsv1.StatefulSet, error) {
	statefulsets, err := GetStatefulSets(kubeClient, sfs.Name)
	if statefulsets != nil {
		//TODO Update statefulsets
		log.Error().Msg("Statefulsets already exist")
		return statefulsets, err
	}

	sfs, err = kubeClient.AppsV1().StatefulSets(namespace).Create(sfs)
	if err != nil {
		//TODO error & logs
		return nil, errors.Wrapf(err, "creating statefulsets '%s'", sfs)
	}

	log.Debug().Msg("Statefulsets created " + sfs.Namespace + "/" + sfs.Name)
	return sfs, nil
}

func DeleteStatefulSets(kubeClient kubernetes.Interface, namespace, id string) error {
	sfs, err := GetStatefulSets(kubeClient, id)
	if sfs == nil {
		errors.Wrapf(err, "Statefulsets not exist")
	}

	deletePolicy := metav1.DeletePropagationForeground
	err = kubeClient.AppsV1().StatefulSets(namespace).Delete(id, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	log.Debug().Msg("Statefulsets deleted " + sfs.namespace + "/" + sfs.Name)
	return err
}

func GetStatefulSetsList(kubeClient kubernetes.Interface, namespace string) ([]string, error) {
	var nodelist []string
	result, err := kubeClient.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: "node",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get statefulsets list '%s'", result)
	}
	for node := 0; node < len(result.Items); node++ {
		nodelist = append(nodelist, result.Items[node].Name)
	}

	log.Debug().Msg("Success Get StatefulSets List")
	return nodelist, nil
}

func GetStatefulSets(kubeClient kubernetes.Interface, namespace string, id string) (*appsv1.StatefulSet, error) {
	sfs, err := kubeClient.AppsV1().StatefulSets(namespace).Get(id, metav1.GetOptions{})

	if err != nil {
		return nil, err // not exist
	}

	log.Debug().Msg("Success Get StatefulSets " + id)
	return sfs, nil // exist
}
