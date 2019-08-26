package statefulsets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var kubernetesInterface *fake.Clientset

var TestReplicas = int32(1)

var TestID = "test"
var TestNamespace = "test"
var TestImage = "foo/bar"
var TestTag = "test"

var testdata = &appsv1.StatefulSet{
	ObjectMeta: metav1.ObjectMeta{
		Name: TestID,
		Labels: map[string]string{
			"node": TestID,
		},
		Namespace: TestNamespace,
	},
	Spec: appsv1.StatefulSetSpec{
		Replicas: &TestReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"node": TestID,
			},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"node": TestID,
				},
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:            TestID,
						Image:           TestImage + ":" + TestTag,
						ImagePullPolicy: "Always",
						Env: []apiv1.EnvVar{
							{
								Name:  "test",
								Value: "config.json",
							},
						},
						Ports: []apiv1.ContainerPort{
							{
								Name:          "test",
								Protocol:      apiv1.ProtocolTCP,
								ContainerPort: 8080,
							},
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      "config",
								MountPath: "test",
							},
						},
					},
				},
				Volumes: []apiv1.Volume{
					{
						Name: "config",
						VolumeSource: apiv1.VolumeSource{
							NFS: &apiv1.NFSVolumeSource{
								Server: "test." + TestNamespace + ".svc.cluster.local",
								Path:   "/" + TestNamespace,
							},
						},
					},
				},
			},
		},
	},
}

func TestMain(m *testing.M) {
	Namespace = TestNamespace
	kubernetesInterface = SetKubernetesInterface()
	os.Exit(m.Run())
}

func SetKubernetesInterface() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestCreateStatefulSets(t *testing.T) {
	_, err := CreateStatefulSets(kubernetesInterface, testdata)

	assert.NoError(t, err, "Error CreateStatefulSets")
}

func TestGetStatefulSets(t *testing.T) {
	result, err := GetStatefulSets(kubernetesInterface, TestID)
	if err != nil {
		assert.NoError(t, err, "Error GetStatefulSets")
	}
	assert.Equal(t, result.Name, TestID)
}

func TestGetStatefulSetsList(t *testing.T) {
	listresult, err := GetStatefulSetsList(kubernetesInterface)
	assert.NoError(t, err, "Error GetStatefulSetsList")
	assert.Equal(t, listresult[0], TestID)
}

func TestDeleteStatefulSets(t *testing.T) {
	err := DeleteStatefulSets(kubernetesInterface, TestID)
	assert.NoError(t, err, "Error DeleteStatefulSets")
}
