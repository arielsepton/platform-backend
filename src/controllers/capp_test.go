package controllers

import (
	"fmt"
	cappv1alpha1 "github.com/dana-team/container-app-operator/api/v1alpha1"
	"github.com/dana-team/platform-backend/src/customerrors"
	"github.com/dana-team/platform-backend/src/types"
	"github.com/dana-team/platform-backend/src/utils/pagination"
	"github.com/dana-team/platform-backend/src/utils/testutils"
	"github.com/dana-team/platform-backend/src/utils/testutils/mocks"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"testing"
)

func TestGetCapp(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-get"

	type requestParams struct {
		name      string
		namespace string
	}

	type want struct {
		capp        types.Capp
		errorStatus metav1.StatusReason
	}

	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedGettingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + "-1",
			},
			want: want{
				capp: types.Capp{
					Metadata: mocks.PrepareCappMetadata(testutils.CappName+"-1", namespaceName),
					Spec:     mocks.PrepareCappSpec(),
					Status:   mocks.PrepareCappStatus(testutils.CappName+"-1", namespaceName, testutils.Domain),
					Labels:   []types.KeyValue{{Key: testutils.LabelKey + "-1", Value: testutils.LabelValue + "-1"}},
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFailGettingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
			},
			want: want{
				capp:        types.Capp{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
		"ShouldFailGettingCappInNonExistingNamespace": {
			requestParams: requestParams{
				namespace: namespaceName + testutils.NonExistentSuffix,
				name:      testutils.CappName,
			},
			want: want{
				capp:        types.Capp{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)
	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.GetCapp(test.requestParams.namespace, test.requestParams.name)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()
				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.capp, response)
		})

	}

}

func TestGetCappState(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-getState"

	type requestParams struct {
		name      string
		namespace string
	}

	type want struct {
		cappState   types.GetCappStateResponse
		errorStatus metav1.StatusReason
	}

	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedGettingEnabledCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, testutils.EnabledState),
			},
			want: want{
				cappState: types.GetCappStateResponse{
					LastCreatedRevision: fmt.Sprintf("%s-%s-%s", testutils.CappName, testutils.EnabledState, "00001"),
					LastReadyRevision:   fmt.Sprintf("%s-%s-%s", testutils.CappName, testutils.EnabledState, "00001"),
					State:               testutils.EnabledState,
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldSucceedGettingDisabledCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, testutils.DisabledState),
			},
			want: want{
				cappState: types.GetCappStateResponse{
					LastCreatedRevision: testutils.NoRevision,
					LastReadyRevision:   testutils.NoRevision,
					State:               testutils.DisabledState,
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFailGettingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
			},
			want: want{
				cappState:   types.GetCappStateResponse{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
		"ShouldFailGettingCappInNonExistingNamespace": {
			requestParams: requestParams{
				namespace: namespaceName + testutils.NonExistentSuffix,
				name:      testutils.CappName,
			},
			want: want{
				cappState:   types.GetCappStateResponse{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)

	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCappWithState(dynClient, fmt.Sprintf("%s-%s", testutils.CappName, testutils.EnabledState),
		namespaceName, testutils.EnabledState, map[string]string{}, map[string]string{})
	mocks.CreateTestCappWithState(dynClient, fmt.Sprintf("%s-%s", testutils.CappName, testutils.DisabledState),
		namespaceName, testutils.DisabledState, map[string]string{}, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.GetCappState(test.requestParams.namespace, test.requestParams.name)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.cappState, response)
		})

	}

}

func TestGetCappDNS(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-getDNS"

	type requestParams struct {
		name      string
		namespace string
	}

	type dnsParams struct {
		readyStatus   corev1.ConditionStatus
		syncedStatus  corev1.ConditionStatus
		isConditioned bool
		hostname      string
	}

	type want struct {
		cappDNS     types.GetDNSResponse
		errorStatus metav1.StatusReason
	}

	cases := map[string]struct {
		requestParams requestParams
		want          want
		records       []dnsParams
		cappName      string
	}{
		"ShouldSucceedGettingAvailableDNS": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, testutils.Available),
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				cappDNS: types.GetDNSResponse{Records: []types.DNS{
					{Status: corev1.ConditionFalse, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
					{Status: corev1.ConditionTrue, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-2", testutils.DefaultZone)},
				},
				},
			},
			records: []dnsParams{
				{readyStatus: corev1.ConditionFalse, syncedStatus: corev1.ConditionTrue, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
				{readyStatus: corev1.ConditionTrue, syncedStatus: corev1.ConditionTrue, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-2", testutils.DefaultZone)},
			},
			cappName: fmt.Sprintf("%s-%s", testutils.CappName, testutils.Available),
		},

		"ShouldSucceedGettingMixedRecords": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, "mixed"),
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				cappDNS: types.GetDNSResponse{Records: []types.DNS{
					{Status: corev1.ConditionFalse, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
					{Status: corev1.ConditionTrue, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-2", testutils.DefaultZone)},
				},
				},
			},
			records: []dnsParams{
				{readyStatus: corev1.ConditionTrue, syncedStatus: corev1.ConditionFalse, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
				{readyStatus: corev1.ConditionTrue, syncedStatus: corev1.ConditionTrue, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-2", testutils.DefaultZone)},
			},
			cappName: fmt.Sprintf("%s-%s", testutils.CappName, "mixed"),
		},
		"ShouldSucceedGettingNotReadyUnknownDNS": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, "ready-unknown"),
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				cappDNS: types.GetDNSResponse{Records: []types.DNS{
					{Status: corev1.ConditionUnknown, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
				},
				},
			},
			records: []dnsParams{
				{readyStatus: corev1.ConditionUnknown, syncedStatus: corev1.ConditionFalse, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
			},
			cappName: fmt.Sprintf("%s-%s", testutils.CappName, "ready-unknown"),
		},

		"ShouldSucceedGettingSyncedUnknownDNS": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      fmt.Sprintf("%s-%s", testutils.CappName, "synced-unknown"),
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				cappDNS: types.GetDNSResponse{Records: []types.DNS{
					{Status: corev1.ConditionUnknown, Name: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
				},
				},
			},
			records: []dnsParams{
				{readyStatus: corev1.ConditionUnknown, syncedStatus: corev1.ConditionTrue, isConditioned: true, hostname: fmt.Sprintf("%s.%s", testutils.Hostname+"-1", testutils.DefaultZone)},
			},
			cappName: fmt.Sprintf("%s-%s", testutils.CappName, "synced-unknown"),
		},

		"ShouldFailGettingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
			},
			want: want{
				cappDNS:     types.GetDNSResponse{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
		"ShouldFailGettingCappInNonExistingNamespace": {
			requestParams: requestParams{
				namespace: namespaceName + testutils.NonExistentSuffix,
				name:      testutils.CappName,
			},
			want: want{
				cappDNS:     types.GetDNSResponse{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)

	createTestNamespace(namespaceName, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {

			if test.cappName != "" {
				mocks.CreateTestCapp(dynClient, test.cappName, namespaceName, testutils.Domain, map[string]string{}, map[string]string{})
			}

			for i, dns := range test.records {
				if !dns.isConditioned {
					mocks.CreateTestCNAMERecordWithoutConditions(dynClient, test.cappName+strconv.Itoa(i), test.cappName, namespaceName, dns.hostname)
				} else {
					mocks.CreateTestCNAMERecord(dynClient, test.cappName+strconv.Itoa(i), test.cappName, namespaceName, dns.hostname, dns.readyStatus, dns.syncedStatus)
				}
			}

			response, err := cappController.GetCappDNS(test.requestParams.namespace, test.requestParams.name)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.cappDNS, response)
		})

	}

}

func TestGetCapps(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-getmany"

	type requestParams struct {
		cappQuery types.CappQuery
		namespace string
		limit     int
		page      int
	}

	type want struct {
		cappList    types.CappList
		errorStatus metav1.StatusReason
	}

	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedGettingAllCappRevisions": {
			requestParams: requestParams{
				namespace: namespaceName,
				cappQuery: types.CappQuery{},
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				cappList: types.CappList{ListMetadata: types.ListMetadata{Count: 2}, Capps: []types.CappSummary{
					mocks.PrepareCappSummary(testutils.CappName+"-1", namespaceName),

					mocks.PrepareCappSummary(testutils.CappName+"-2", namespaceName),
				},
				},
			},
		},
		"ShouldSucceedGettingCappByLabels": {
			requestParams: requestParams{
				namespace: namespaceName,
				cappQuery: types.CappQuery{LabelSelector: fmt.Sprintf("%s-2=%s-2", testutils.LabelKey, testutils.LabelValue)},
			},
			want: want{
				cappList: types.CappList{ListMetadata: types.ListMetadata{Count: 1}, Capps: []types.CappSummary{
					mocks.PrepareCappSummary(testutils.CappName+"-2", namespaceName),
				},
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFailGettingCappsWithInvalidSelector": {
			requestParams: requestParams{
				namespace: namespaceName,
				cappQuery: types.CappQuery{LabelSelector: testutils.InvalidLabelSelector},
			},
			want: want{
				cappList:    types.CappList{},
				errorStatus: metav1.StatusReasonBadRequest,
			},
		},
		"ShouldFailGettingNonExistingNamespace": {
			requestParams: requestParams{
				namespace: namespaceName + testutils.NonExistentSuffix,
				cappQuery: types.CappQuery{},
			},
			want: want{
				cappList:    types.CappList{},
				errorStatus: metav1.StatusSuccess,
			},
		},
	}
	setup()

	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-2", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-2": testutils.LabelValue + "-2"}, map[string]string{})
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			c := mocks.GinContext()
			mocks.SetPaginationValues(c, test.requestParams.limit, test.requestParams.page)
			cappController := NewCappController(dynClient, c, logger)

			limit, page, _ := pagination.ExtractPaginationParamsFromCtx(c)
			response, err := cappController.GetCapps(test.requestParams.namespace, limit, page, test.requestParams.cappQuery)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.cappList, response)
		})
	}
}

func TestCreateCapp(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-create"
	type requestParams struct {
		capp      types.CreateCapp
		namespace string
	}

	type want struct {
		response    types.Capp
		errorStatus metav1.StatusReason
	}
	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedCreatingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				capp:      mocks.PrepareCreateCappType(testutils.CappName+"-2", []types.KeyValue{{Key: testutils.LabelKey + "-2", Value: testutils.LabelValue + "-2"}}, nil),
			},
			want: want{
				response: types.Capp{
					Metadata: mocks.PrepareCappMetadata(testutils.CappName+"-2", namespaceName),
					Spec:     mocks.PrepareCappSpec(),
					Status:   cappv1alpha1.CappStatus{},
					Labels:   []types.KeyValue{{Key: testutils.LabelKey + "-2", Value: testutils.LabelValue + "-2"}},
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFailCreatingExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				capp:      mocks.PrepareCreateCappType(testutils.CappName+"-1", []types.KeyValue{{Key: testutils.LabelKey + "-1", Value: testutils.LabelValue + "-1"}}, nil),
			},
			want: want{
				response:    types.Capp{},
				errorStatus: metav1.StatusReasonAlreadyExists,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)
	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.CreateCapp(test.requestParams.namespace, test.requestParams.capp)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.response, response)
		})
	}
}

func TestUpdateCapp(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-update"
	type requestParams struct {
		name      string
		capp      types.UpdateCapp
		namespace string
	}

	type want struct {
		response    types.Capp
		errorStatus metav1.StatusReason
	}
	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedUpdatingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + "-1",
				capp:      mocks.PrepareUpdateCappType([]types.KeyValue{{Key: testutils.LabelKey + "-3", Value: testutils.LabelValue + "-3"}}, nil),
			},
			want: want{
				response: types.Capp{
					Metadata: mocks.PrepareCappMetadata(testutils.CappName+"-1", namespaceName),
					Spec:     mocks.PrepareCappSpec(),
					Status:   mocks.PrepareCappStatus(testutils.CappName+"-1", namespaceName, testutils.Domain),
					Labels:   []types.KeyValue{{Key: testutils.LabelKey + "-3", Value: testutils.LabelValue + "-3"}},
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFaildUpdatingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
				capp:      mocks.PrepareUpdateCappType([]types.KeyValue{{Key: testutils.LabelKey + "-3", Value: testutils.LabelValue + "-3"}}, nil),
			},
			want: want{
				response:    types.Capp{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)
	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.UpdateCapp(test.requestParams.namespace, test.requestParams.name, test.requestParams.capp)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.response, response)
		})
	}
}

func TestEditCapp(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-update"
	type requestParams struct {
		name      string
		state     string
		namespace string
	}

	type want struct {
		response    types.CappStateReponse
		errorStatus metav1.StatusReason
	}
	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedEditingCappState": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + "-1",
				state:     testutils.DisabledState,
			},
			want: want{
				response: types.CappStateReponse{
					State: testutils.DisabledState,
					Name:  testutils.CappName + "-1",
				},
				errorStatus: metav1.StatusSuccess,
			},
		},
		"ShouldFailedEditingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
				state:     testutils.DisabledState,
			},
			want: want{
				response:    types.CappStateReponse{},
				errorStatus: metav1.StatusReasonNotFound,
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)
	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.EditCappState(test.requestParams.namespace, test.requestParams.name, test.requestParams.state)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.response, response)
		})
	}
}

func TestDeleteCapp(t *testing.T) {
	namespaceName := testutils.CappNamespace + "-delete"
	type requestParams struct {
		name      string
		namespace string
	}
	type want struct {
		response    types.CappError
		errorStatus metav1.StatusReason
	}
	cases := map[string]struct {
		requestParams requestParams
		want          want
	}{
		"ShouldSucceedDeletingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + "-1",
			},
			want: want{
				errorStatus: metav1.StatusSuccess,
				response: types.CappError{
					Message: fmt.Sprintf("Deleted capp %q in namespace %q successfully", testutils.CappName+"-1", namespaceName),
				},
			},
		},
		"ShouldFailDeletingNonExistingCapp": {
			requestParams: requestParams{
				namespace: namespaceName,
				name:      testutils.CappName + testutils.NonExistentSuffix,
			},
			want: want{
				errorStatus: metav1.StatusReasonNotFound,
				response:    types.CappError{},
			},
		},
	}
	setup()
	cappController := NewCappController(dynClient, mocks.GinContext(), logger)
	createTestNamespace(namespaceName, map[string]string{})
	mocks.CreateTestCapp(dynClient, testutils.CappName+"-1", namespaceName, testutils.Domain, map[string]string{testutils.LabelKey + "-1": testutils.LabelValue + "-1"}, map[string]string{})

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			response, err := cappController.DeleteCapp(test.requestParams.namespace, test.requestParams.name)
			if test.want.errorStatus != metav1.StatusSuccess {
				reason := err.(customerrors.ErrorWithStatusCode).StatusReason()

				assert.Equal(t, test.want.errorStatus, reason)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want.response, response)
		})
	}
}
