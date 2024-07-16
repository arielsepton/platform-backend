package mocks

import (
	"github.com/dana-team/platform-backend/src/utils/testutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PreparePod simulates creating a pod and adding some log lines.
func PreparePod(namespace, podName, cappName string) *corev1.Pod {
	labels := map[string]string{}

	if cappName != "" {
		labels = map[string]string{
			testutils.CappResourceKey: cappName,
		}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx",
				},
				{
					Name:  "test-capp",
					Image: "nginx",
				},
			},
		},
	}
}
