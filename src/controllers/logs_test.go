package controllers

import (
	"context"
	"github.com/dana-team/platform-backend/src/utils/testutils/mocks"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
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

func TestFetchPodLogs(t *testing.T) {
	type args struct {
		client        kubernetes.Interface
		namespace     string
		podName       string
		containerName string
	}
	type want struct {
		errContains string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"ShouldFailGettingLogsOnNonExistingPod": {
			args: args{
				client:        fakeClient,
				namespace:     "test-namespace",
				podName:       "non-existent-pod",
				containerName: "test-container",
			},
			want: want{
				errContains: "error opening log stream",
			},
		},
	}

	setup()
	mocks.CreateTestPod(fakeClient, "test-namespace", "test-pod-1", "")
	mockLogger, _ := zap.NewDevelopment()

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := FetchPodLogs(context.TODO(), tc.args.client, tc.args.namespace, tc.args.podName, tc.args.containerName, mockLogger)
			if tc.want.errContains == "" {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.want.errContains) {
					t.Errorf("Expected error containing '%s', but got: %v", tc.want.errContains, err)
				}
			}
		})
	}
}

func TestFetchCappLogs(t *testing.T) {
	type args struct {
		client        kubernetes.Interface
		namespace     string
		cappName      string
		containerName string
		podName       string
	}
	type want struct {
		errContains string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"ShouldReturnNoPodsForNonExistingCapp": {
			args: args{
				client:        fakeClient,
				namespace:     "test-namespace",
				cappName:      "non-existent-capp",
				containerName: "test-container",
				podName:       "test-pod-1",
			},
			want: want{
				errContains: "no pods found for Capp",
			},
		},
		"ShouldFailOnNonExistingPod": {
			args: args{
				client:        fakeClient,
				namespace:     "test-namespace",
				cappName:      "test-capp",
				containerName: "test-container",
				podName:       "non-existent-pod",
			},
			want: want{
				errContains: "o pods found for Capp test-capp in namespace test-namespace",
			},
		},
	}

	setup()
	mocks.CreateTestPod(fakeClient, "test-namespace", "test-pod-1", "test-capp")
	mockLogger, _ := zap.NewDevelopment()

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := FetchCappLogs(context.TODO(), tc.args.client, tc.args.namespace, tc.args.cappName, tc.args.containerName, tc.args.podName, mockLogger)
			if tc.want.errContains == "" {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.want.errContains) {
					t.Errorf("Expected error containing '%s', but got: %v", tc.want.errContains, err)
				}
			}
		})
	}
}
