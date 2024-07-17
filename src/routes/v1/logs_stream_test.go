package v1

import (
	"github.com/dana-team/platform-backend/src/middleware"
	"github.com/dana-team/platform-backend/src/utils/testutils"
	"github.com/dana-team/platform-backend/src/utils/testutils/mocks"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testNamespace       = testutils.TestNamespace + "-logs"
	authorizationHeader = "Authorization"
)

func Test_GetCappLogs(t *testing.T) {
	type args struct {
		token string
		wsUrl string
	}
	type want struct {
		statusCode    int
		expectedLines []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"ShouldStreamLogsWithoutQueryParams": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/test-capp",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
		"ShouldNotStreamLogsWithInvalidCappName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/invalid-capp",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming \"Capp\" logs: no pods found for Capp invalid-capp in namespace " + testNamespace},
			},
		},
		"ShouldStreamLogsWithQueryParams": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/test-capp?podName=test-pod-2",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
		"ShouldNotStreamLogsWithNonExistingPodName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/test-capp?podName=pod" + testutils.NonExistentSuffix,
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming \"Capp\" logs: pod 'pod" + testutils.NonExistentSuffix + "' not found for Capp test-capp in namespace " + testNamespace},
			},
		},
		"ShouldNotStreamLogsWithInvalidContainerName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/test-capp?container=container" + testutils.NonExistentSuffix,
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming \"Capp\" logs: error opening log stream: container container" + testutils.NonExistentSuffix + " not found in the pod test-pod-2"},
			},
		},
		"ShouldStreamLogsWithValidContainerName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/capp/" + testNamespace + "/test-capp?container=test-container",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
	}

	setup()
	mocks.CreateTestPod(fakeClient, testNamespace, "test-pod-1", "", false)
	mocks.CreateTestPod(fakeClient, testNamespace, "test-pod-2", "test-capp", true)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			token = tc.args.token
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + tc.args.wsUrl

			dialer := websocket.DefaultDialer
			headers := http.Header{}
			headers.Add(authorizationHeader, tc.args.token)
			headers.Add(middleware.WebsocketTokenHeader, tc.args.token)

			conn, resp, err := dialer.Dial(wsURL, headers)
			assert.Equal(t, tc.want.statusCode, resp.StatusCode)
			if tc.want.statusCode == http.StatusUnauthorized {
				return
			}

			if err != nil {
				t.Fatalf("Failed to dial WebSocket: %v", err)
			}

			defer conn.Close()
			for _, expectedLine := range tc.want.expectedLines {
				_, message, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("Error reading message from WebSocket: %v", err)
				}
				assert.Contains(t, string(message), expectedLine)
			}
		})
	}
}

func Test_GetPodLogs(t *testing.T) {
	type args struct {
		token string
		wsUrl string
	}
	type want struct {
		statusCode    int
		expectedLines []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"ShouldStreamLogsWithoutQueryParams": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/pod/" + testNamespace + "/test-pod-1",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Pod: test-pod-1 line: fake logs"},
			},
		},
		"ShouldStreamLogsWithQueryParams": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/pod/" + testNamespace + "/test-pod-1?container=test-container",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Pod: test-pod-1 line: fake logs"},
			},
		},
		"ShouldStreamLogsWithoutQueryParamsMultipleContainers": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/pod/" + testNamespace + "/test-pod-3",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming \"Pod\" logs: error opening log stream: pod test-pod-3 has multiple containers, please specify the container name"},
			},
		},
		"ShouldNotStreamLogsWithNonExistingPodName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/pod/" + testNamespace + "/test-invalid-pod",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{`error: Error streaming "Pod" logs: error opening log stream: failed to get pod: pods "test-invalid-pod" not found`},
			},
		},
		"ShouldNotStreamLogsWithInvalidContainerName": {
			args: args{
				token: "valid_token",
				wsUrl: "/v1/logs/pod/" + testNamespace + "/test-pod-1?container=container" + testutils.NonExistentSuffix,
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming \"Pod\" logs: error opening log stream: container container" + testutils.NonExistentSuffix + " not found in the pod test-pod-1"},
			},
		},
	}

	setup()
	mocks.CreateTestPod(fakeClient, testNamespace, "test-pod-1", "", false)
	mocks.CreateTestPod(fakeClient, testNamespace, "test-pod-2", "test-capp", false)
	mocks.CreateTestPod(fakeClient, testNamespace, "test-pod-3", "test-capp", true)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			token = tc.args.token
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + tc.args.wsUrl

			dialer := websocket.DefaultDialer
			headers := http.Header{}
			headers.Add(authorizationHeader, tc.args.token)
			headers.Add(middleware.WebsocketTokenHeader, tc.args.token)

			conn, resp, err := dialer.Dial(wsURL, headers)
			assert.Equal(t, tc.want.statusCode, resp.StatusCode)
			if tc.want.statusCode == http.StatusUnauthorized {
				return
			}

			if err != nil {
				t.Fatalf("Failed to dial WebSocket: %v", err)
			}

			defer conn.Close()
			for _, expectedLine := range tc.want.expectedLines {
				_, message, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("Error reading message from WebSocket: %v", err)
				}
				assert.Contains(t, string(message), expectedLine)
			}
		})
	}
}
