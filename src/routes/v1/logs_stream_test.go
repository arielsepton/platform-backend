package v1_test

import (
	"github.com/dana-team/platform-backend/src/middleware"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetCappLogs(t *testing.T) {
	type args struct {
		token     string
		namespace string
		cappName  string
		container string
		wsUrl     string
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
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/test-capp",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
		"ShouldNotStreamLogsWithInvalidCappName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/invalid-capp",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming Capp logs: no pods found for Capp invalid-capp in namespace test-namespace"},
			},
		},
		"ShouldStreamLogsWithQueryParams": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/test-capp?podName=test-pod-2",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
		"ShouldNotStreamLogsWithNonExistingPodName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/test-capp?podName=fakepod",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming Capp logs: pod 'fakepod' not found for Capp test-capp in namespace test-namespace"},
			},
		},
		"ShouldNotStreamLogsWithInvalidContainerName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/test-capp?container=nonexisting",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming Capp logs: error opening log stream: container nonexisting not found in the pod test-pod-2"},
			},
		},
		"ShouldStreamLogsWithValidContainerName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				cappName:  "test-capp",
				container: "test-container",
				wsUrl:     "/v1/logs/capp/" + testNamespace + "/test-capp?container=test-container",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Capp: test-capp line: fake logs"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			token = tc.args.token
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + tc.args.wsUrl

			dialer := websocket.DefaultDialer
			headers := http.Header{}
			headers.Add("Authorization", tc.args.token)
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
		token     string
		namespace string
		podName   string
		container string
		wsUrl     string
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
				token:     "valid_token",
				namespace: testNamespace,
				podName:   "test-pod-1",
				container: "test-container",
				wsUrl:     "/v1/logs/pod/" + testNamespace + "/test-pod-1",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Pod: test-pod-1 line: fake logs"},
			},
		},
		"ShouldStreamLogsWithQueryParams": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				podName:   "test-pod-1",
				container: "test-container",
				wsUrl:     "/v1/logs/pod/" + testNamespace + "/test-pod-1?container=test-container",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"Pod: test-pod-1 line: fake logs"},
			},
		},
		"ShouldNotStreamLogsWithNonExistingPodName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				podName:   "test-pod-1",
				container: "test-container",
				wsUrl:     "/v1/logs/pod/" + testNamespace + "/test-invalid-pod",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{`error: Error streaming Pod logs: error opening log stream: failed to get pod: pods "test-invalid-pod" not found`},
			},
		},
		"ShouldNotStreamLogsWithInvalidContainerName": {
			args: args{
				token:     "valid_token",
				namespace: testNamespace,
				podName:   "test-pod-1",
				container: "test-container",
				wsUrl:     "/v1/logs/pod/" + testNamespace + "/test-pod-1?container=non-existing-container",
			},
			want: want{
				statusCode:    http.StatusSwitchingProtocols,
				expectedLines: []string{"error: Error streaming Pod logs: error opening log stream: container non-existing-container not found in the pod test-pod-1"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			token = tc.args.token
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + tc.args.wsUrl

			dialer := websocket.DefaultDialer
			headers := http.Header{}
			headers.Add("Authorization", tc.args.token)
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
