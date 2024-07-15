package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestFetchCappPodName(t *testing.T) {
	type args struct {
		podName string
		pods    *corev1.PodList
	}
	type want struct {
		name  string
		found bool
	}

	// Mock PodList for testing
	mockPodList := &corev1.PodList{
		Items: []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "pod3"}},
		},
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Empty pod name returns first pod": {
			args: args{
				podName: "",
				pods:    mockPodList,
			},
			want: want{
				name:  "pod1",
				found: true,
			},
		},
		"Existing pod name is found": {
			args: args{
				podName: "pod2",
				pods:    mockPodList,
			},
			want: want{
				name:  "pod2",
				found: true,
			},
		},
		"Non-existing pod name returns false": {
			args: args{
				podName: "nonexistent",
				pods:    mockPodList,
			},
			want: want{
				name:  "nonexistent",
				found: false,
			},
		},
		"Empty pod list returns false": {
			args: args{
				podName: "pod1",
				pods:    &corev1.PodList{},
			},
			want: want{
				name:  "pod1",
				found: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualName, actualFound := FetchCappPodName(tc.args.podName, tc.args.pods)
			if actualName != tc.want.name {
				t.Errorf("Expected pod name %s, but got %s", tc.want.name, actualName)
			}
			if actualFound != tc.want.found {
				t.Errorf("Expected found %v, but got %v", tc.want.found, actualFound)
			}
		})
	}
}

//func TestFetchPodLogs(t *testing.T) {
//	// Mock Kubernetes client
//	mockClient := &MockKubernetesClient{}
//
//	type args struct {
//		ctx           context.Context
//		client        kubernetes.Interface
//		namespace     string
//		podName       string
//		containerName string
//	}
//	type want struct {
//		errContains string
//	}
//
//	cases := map[string]struct {
//		args args
//		want want
//	}{
//		"Successful log stream": {
//			args: args{
//				ctx:           context.TODO(),
//				client:        mockClient,
//				namespace:     "test-namespace",
//				podName:       "test-pod-1",
//				containerName: "test-container",
//			},
//			want: want{
//				errContains: "",
//			},
//		},
//		"Error opening log stream": {
//			args: args{
//				ctx:           context.TODO(),
//				client:        mockClient,
//				namespace:     "test-namespace",
//				podName:       "non-existent-pod",
//				containerName: "test-container",
//			},
//			want: want{
//				errContains: "error opening log stream",
//			},
//		},
//	}
//
//	for name, tc := range cases {
//		t.Run(name, func(t *testing.T) {
//			_, err := FetchPodLogs(tc.args.ctx, tc.args.client, tc.args.namespace, tc.args.podName, tc.args.containerName)
//			if tc.want.errContains == "" {
//				if err != nil {
//					t.Errorf("Unexpected error: %v", err)
//				}
//			} else {
//				if err == nil || !strings.Contains(err.Error(), tc.want.errContains) {
//					t.Errorf("Expected error containing '%s', but got: %v", tc.want.errContains, err)
//				}
//			}
//		})
//	}
//}
//
//func TestFetchCappLogs(t *testing.T) {
//	// Mock Kubernetes client
//	mockClient := &MockKubernetesClient{}
//
//	type args struct {
//		ctx           context.Context
//		client        kubernetes.Interface
//		namespace     string
//		cappName      string
//		containerName string
//		podName       string
//	}
//	type want struct {
//		errContains string
//	}
//
//	cases := map[string]struct {
//		args args
//		want want
//	}{
//		"Successful log stream": {
//			args: args{
//				ctx:           context.TODO(),
//				client:        mockClient,
//				namespace:     "test-namespace",
//				cappName:      "test-capp",
//				containerName: "test-container",
//				podName:       "test-pod-1",
//			},
//			want: want{
//				errContains: "",
//			},
//		},
//		"No pods found for Capp": {
//			args: args{
//				ctx:           context.TODO(),
//				client:        mockClient,
//				namespace:     "test-namespace",
//				cappName:      "non-existent-capp",
//				containerName: "test-container",
//				podName:       "test-pod-1",
//			},
//			want: want{
//				errContains: "no pods found for Capp",
//			},
//		},
//		"Pod not found for Capp": {
//			args: args{
//				ctx:           context.TODO(),
//				client:        mockClient,
//				namespace:     "test-namespace",
//				cappName:      "test-capp",
//				containerName: "test-container",
//				podName:       "non-existent-pod",
//			},
//			want: want{
//				errContains: "pod 'non-existent-pod' not found for Capp",
//			},
//		},
//	}
//
//	for name, tc := range cases {
//		t.Run(name, func(t *testing.T) {
//			_, err := FetchCappLogs(tc.args.ctx, tc.args.client, tc.args.namespace, tc.args.cappName, tc.args.containerName, tc.args.podName)
//			if tc.want.errContains == "" {
//				if err != nil {
//					t.Errorf("Unexpected error: %v", err)
//				}
//			} else {
//				if err == nil || !strings.Contains(err.Error(), tc.want.errContains) {
//					t.Errorf("Expected error containing '%s', but got: %v", tc.want.errContains, err)
//				}
//			}
//		})
//	}
//}
