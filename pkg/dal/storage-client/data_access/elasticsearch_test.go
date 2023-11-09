package data_access

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/mocks"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8saudit "k8s.io/apiserver/pkg/apis/audit"
)

func TestProvideEsStorage(t *testing.T) {
	opts := common.NewEsOptions()
	_, err := ProvideEsStorage(opts)
	// 默认配置并不能初始化client
	assert.NotNil(t, err)
}

func TestQuerySpanWithPodUid(t *testing.T) {

	type TestCase struct {
		name           string
		uid            string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchPodWithUid",
			uid:  "abcdef",
			esReqRepExpect: func() map[string]string {
				sp := model.Span{
					Name: "abcdef",
				}
				pb, _ := json.Marshal(sp)
				spBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: spBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"OwnerRef.UID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchPodWithUid",
			uid:  "",
			esReqRepExpect: func() map[string]string {
				sp := model.Span{
					Name: "abcdef",
				}
				pb, _ := json.Marshal(sp)
				spBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: spBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"OwnerRef.UID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchPodWithUid",
			uid:  "123",
			esReqRepExpect: func() map[string]string {
				sp := model.Span{
					Name: "abcdef",
				}
				pb, _ := json.Marshal(sp)
				spBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: spBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"abc": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()

		client, err := mock.MockElasticSearchClient()

		storageEsImpl := &StorageEsImpl{
			DB: client,
		}

		res := make([]model.Span, 0)
		if tc.uid == "abcdef" {
			err = storageEsImpl.QuerySpanWithPodUid(&res, tc.uid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		}

		if tc.uid == "" || tc.uid == "123" {
			err = storageEsImpl.QuerySpanWithPodUid(&res, tc.uid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QuerySpanWithPodUid(&resErr, tc.uid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryLifePhaseWithPodUid(t *testing.T) {

	type TestCase struct {
		name           string
		uid            string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchLifePhaseWithUid",
			uid:  "abcdef",
			esReqRepExpect: func() map[string]string {
				lifePhase := model.LifePhase{
					PodUID: "abcdef",
				}
				lp, _ := json.Marshal(lifePhase)
				phaseBytes := json.RawMessage(lp)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: phaseBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchLifePhaseWithUid",
			uid:  "",
			esReqRepExpect: func() map[string]string {
				lifePhase := model.LifePhase{
					PodUID: "abcdef",
				}
				lp, _ := json.Marshal(lifePhase)
				phaseBytes := json.RawMessage(lp)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: phaseBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.LifePhase, 0)
		if tc.uid != "" {
			err = storageEsImpl.QueryLifePhaseWithPodUid(&res, tc.uid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryLifePhaseWithPodUid(&res, tc.uid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryLifePhaseWithPodUid(&resErr, tc.uid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestQueryPodYamlsWithPodUID(t *testing.T) {

	type TestCase struct {
		name           string
		uid            string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchPodWithUid",
			uid:  "abcdef",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchPodWithUid",
			uid:  "",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.uid != "" {
			err = storageEsImpl.QueryPodYamlsWithPodUID(&res, tc.uid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlsWithPodUID(&res, tc.uid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlsWithPodUID(&resErr, tc.uid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryPodYamlsWithPodName(t *testing.T) {

	type TestCase struct {
		name           string
		podName        string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:    "FetchPodWithUid",
			podName: "abcdef",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:    "FetchPodWithUid",
			podName: "",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.podName != "" {
			err = storageEsImpl.QueryPodYamlsWithPodName(&res, tc.podName)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlsWithPodName(&res, tc.podName)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlsWithPodName(&resErr, tc.podName)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestQueryPodYamlsWithHostName(t *testing.T) {

	type TestCase struct {
		name           string
		Hostname       string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:     "FetchPodWithHostname",
			Hostname: "abcdef",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					Hostname: "abcdef",
					Pod:      &v1.Pod{Spec: v1.PodSpec{Hostname: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostname": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:     "FetchPodWithHostname",
			Hostname: "",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					Hostname: "abcdef",
					Pod:      &v1.Pod{Spec: v1.PodSpec{Hostname: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostname": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.Hostname != "" {
			err = storageEsImpl.QueryPodYamlsWithHostName(&res, tc.Hostname)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlsWithHostName(&res, tc.Hostname)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlsWithHostName(&resErr, tc.Hostname)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryPodYamlsWithPodIp(t *testing.T) {

	type TestCase struct {
		name           string
		PodIP          string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:  "FetchPodWithPodip",
			PodIP: "12.34.56.78",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodIP: "abcdef",
					Pod:   &v1.Pod{Status: v1.PodStatus{PodIP: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podIP": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:  "FetchPodWithPodip",
			PodIP: "",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodIP: "abcdef",
					Pod:   &v1.Pod{Status: v1.PodStatus{PodIP: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podIP": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.PodIP != "" {
			err = storageEsImpl.QueryPodYamlsWithPodIp(&res, tc.PodIP)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlsWithPodIp(&res, tc.PodIP)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlsWithPodIp(&resErr, tc.PodIP)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryPodListWithNodeip(t *testing.T) {

	type TestCase struct {
		name           string
		nodeIp         string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchPodWithNodeip",
			nodeIp: "abcdef",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					HostIP: "abcdef",
					PodUid: "12345",
					Pod:    &v1.Pod{Status: v1.PodStatus{HostIP: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchPodWithNodeip",
			nodeIp: "",
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					HostIP: "abcdef",
					PodUid: "12345",
					Pod:    &v1.Pod{Status: v1.PodStatus{HostIP: "abcdef"}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.nodeIp != "" {
			err = storageEsImpl.QueryPodListWithNodeip(&res, tc.nodeIp, false)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodListWithNodeip(&res, tc.nodeIp, false)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodListWithNodeip(&resErr, tc.nodeIp, false)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func Test_getPodUIDListByHostname(t *testing.T) {

	type TestCase struct {
		name           string
		hostname       string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:     "queryPodUIDListByHostname",
			hostname: "abcdef",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostname.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:     "queryPodUIDListByHostname",
			hostname: "",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostname.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.hostname != "" {
			err = storageEsImpl.QueryPodUIDListByHostname(&res, tc.hostname)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodUIDListByHostname(&res, tc.hostname)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodUIDListByHostname(&resErr, tc.hostname)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func Test_getPodUIDListByPodName(t *testing.T) {

	type TestCase struct {
		name           string
		podName        string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:    "queryPodUIDListByPodName",
			podName: "abcdef",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:    "queryPodUIDListByPodName",
			podName: "",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}
	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.podName != "" {
			err = storageEsImpl.QueryPodUIDListByPodName(&res, tc.podName)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodUIDListByPodName(&res, tc.podName)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodUIDListByPodName(&resErr, tc.podName)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func Test_getPodUIDListByPodIP(t *testing.T) {

	type TestCase struct {
		name           string
		ip             string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "queryPodUIDListByPodName",
			ip:   "abcdef",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "queryPodUIDListByPodName",
			ip:   "",
			esReqRepExpect: func() map[string]string {
				var res = "{\"podUID\": \"abcdef\"}"
				sloBytes := json.RawMessage(res)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}
	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.ip != "" {
			err = storageEsImpl.QueryPodUIDListByPodIP(&res, tc.ip)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodUIDListByPodIP(&res, tc.ip)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodUIDListByPodIP(&resErr, tc.ip)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestQueryNodeYamlsWithNodeName(t *testing.T) {

	type TestCase struct {
		name           string
		nodeName       string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:     "FetchNodeWithNodename",
			nodeName: "abcdef",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				ny, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:     "FetchNodeWithNodename",
			nodeName: "",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				ny, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.NodeYaml, 0)
		if tc.nodeName != "" {
			err = storageEsImpl.QueryNodeYamlsWithNodeName(&res, tc.nodeName)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodeYamlsWithNodeName(&res, tc.nodeName)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodeYamlsWithNodeName(&resErr, tc.nodeName)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryNodeYamlsWithNodeUid(t *testing.T) {

	type TestCase struct {
		name           string
		uid            string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchNodeWithNodename",
			uid:  "abcdef",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"uid": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchNodeWithNodename",
			uid:  "",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"uid": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.NodeYaml, 0)
		if tc.uid != "" {
			err = storageEsImpl.QueryNodeYamlsWithNodeUid(&res, tc.uid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodeYamlsWithNodeUid(&res, tc.uid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodeYamlsWithNodeUid(&resErr, tc.uid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryNodeUIDListWithNodeIp(t *testing.T) {

	type TestCase struct {
		name           string
		nodeIp         string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "queryNodeUIDListByNodeIp",
			nodeIp: "abcdef",
			esReqRepExpect: func() map[string]string {
				var res = "{\"uid\": \"abcdef\"}"
				sloBytes := json.RawMessage([]byte(res))
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeIp": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "queryNodeUIDListByNodeIp",
			nodeIp: "",
			esReqRepExpect: func() map[string]string {
				var res = "{\"uid\": \"abcdef\"}"
				sloBytes := json.RawMessage([]byte(res))
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeIp": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.NodeYaml, 0)
		if tc.nodeIp != "" {
			err = storageEsImpl.QueryNodeUIDListWithNodeIp(&res, tc.nodeIp)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodeUIDListWithNodeIp(&res, tc.nodeIp)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodeUIDListWithNodeIp(&resErr, tc.nodeIp)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryNodeYamlsWithNodeIP(t *testing.T) {

	type TestCase struct {
		name           string
		nodeIp         string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchNodeWithNodeIp",
			nodeIp: "abcdef",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeIp": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchNodeWithNodeIp",
			nodeIp: "",
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				pb, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeIp": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.NodeYaml, 0)
		if tc.nodeIp != "" {
			err = storageEsImpl.QueryNodeYamlsWithNodeIP(&res, tc.nodeIp)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodeYamlsWithNodeIP(&res, tc.nodeIp)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodeYamlsWithNodeIP(&resErr, tc.nodeIp)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryPodInfoWithPodUid(t *testing.T) {

	type TestCase struct {
		name           string
		uid            string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchPodWithUid",
			uid:  "abcdef",
			esReqRepExpect: func() map[string]string {
				podInfo := model.PodInfo{
					PodUID: "abcdef",
				}
				pb, _ := json.Marshal(podInfo)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name: "FetchPodWithUid",
			uid:  "",
			esReqRepExpect: func() map[string]string {
				podInfo := model.PodInfo{
					PodUID: "abcdef",
				}
				pb, _ := json.Marshal(podInfo)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodInfo, 0)
		if tc.uid != "" {
			err = storageEsImpl.QueryPodInfoWithPodUid(&res, tc.uid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodInfoWithPodUid(&res, tc.uid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodInfoWithPodUid(&resErr, tc.uid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestQueryPodYamlsWithNodeIP(t *testing.T) {

	type TestCase struct {
		name           string
		nodeIp         string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchPodYamlWithNodeIp",
			nodeIp: "abcdef",
			esReqRepExpect: func() map[string]string {
				podYaml := &model.PodYaml{
					Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(time.Now())}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchPodYamlWithNodeIp",
			nodeIp: "",
			esReqRepExpect: func() map[string]string {
				podYaml := &model.PodYaml{
					Pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(time.Now())}},
				}
				pb, _ := json.Marshal(podYaml)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"hostIP.keyword": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.PodYaml, 0)
		if tc.nodeIp != "" {
			err = storageEsImpl.QueryPodYamlsWithNodeIP(&res, tc.nodeIp)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlsWithNodeIP(&res, tc.nodeIp)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlsWithNodeIP(&resErr, tc.nodeIp)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryNodephaseWithNodeName(t *testing.T) {

	type TestCase struct {
		name           string
		nodeName       string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:     "FetchNodePhaseWithNodeUid",
			nodeName: "abcdef",
			esReqRepExpect: func() map[string]string {
				lifePhase := &model.NodeLifePhase{PlfID: "abcdef"}
				pb, _ := json.Marshal(lifePhase)
				lifeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: lifeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:     "FetchNodePhaseWithNodeUid",
			nodeName: "",
			esReqRepExpect: func() map[string]string {
				lifePhase := &model.NodeLifePhase{PlfID: "abcdef"}
				pb, _ := json.Marshal(lifePhase)
				lifeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: lifeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.NodeLifePhase, 0)
		if tc.nodeName != "" {
			err = storageEsImpl.QueryNodephaseWithNodeName(&res, tc.nodeName)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodephaseWithNodeName(&res, tc.nodeName)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodephaseWithNodeName(&resErr, tc.nodeName)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryNodephaseWithNodeUID(t *testing.T) {

	type TestCase struct {
		name           string
		nodeUid        string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:    "FetchNodePhaseWithNodeUid",
			nodeUid: "abcdef",
			esReqRepExpect: func() map[string]string {
				lifePhase := &model.NodeLifePhase{PlfID: "abcdef"}
				pb, _ := json.Marshal(lifePhase)
				lifeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: lifeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"uid": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:    "FetchNodePhaseWithNodeUid",
			nodeUid: "",
			esReqRepExpect: func() map[string]string {
				lifePhase := &model.NodeLifePhase{PlfID: "abcdef"}
				pb, _ := json.Marshal(lifePhase)
				lifeBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: lifeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"uid": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.NodeLifePhase, 0)
		if tc.nodeUid != "" {
			err = storageEsImpl.QueryNodephaseWithNodeUID(&res, tc.nodeUid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodephaseWithNodeUID(&res, tc.nodeUid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodephaseWithNodeUID(&resErr, tc.nodeUid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func TestQueryDebuggingWithPodUid(t *testing.T) {

	type TestCase struct {
		name           string
		podUid         string
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "querySloTraceDataWithPodUID",
			podUid: "abcdef",
			esReqRepExpect: func() map[string]string {
				slotrace := &model.SloTraceData{}
				pb, _ := json.Marshal(slotrace)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"PodUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "querySloTraceDataWithPodUID",
			podUid: "",
			esReqRepExpect: func() map[string]string {
				slotrace := &model.SloTraceData{}
				pb, _ := json.Marshal(slotrace)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"PodUID": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.SloTraceData, 0)
		if tc.podUid != "" {
			err = storageEsImpl.QuerySloTraceDataWithPodUID(&res, tc.podUid)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QuerySloTraceDataWithPodUID(&res, tc.podUid)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QuerySloTraceDataWithPodUID(&resErr, tc.podUid)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func Test_querySloByResult(t *testing.T) {

	type TestCase struct {
		name           string
		params         *model.SloOptions
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchquerySloByResultProd",
			params: &model.SloOptions{SloTime: "10s", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Env: "prod"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"SLOViolationReason": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchquerySloByResultTest",
			params: &model.SloOptions{Count: "5", SloTime: "10s", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Env: "test"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"SLOViolationReason": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchquerySloByResultDryrun",
			params: &model.SloOptions{SloTime: "10s", DeliveryStatus: "success", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Env: "dryrun"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"SLOViolationReason": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchquerySloByResultDryrun",
			params: &model.SloOptions{SloTime: "10s", BizName: "biz", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Env: "dryrun", Cluster: "test"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"SLOViolationReason": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchquerySloByResultDryrun",
			params: &model.SloOptions{SloTime: "10s", BizName: "biz", Result: "", From: time.Now(), To: time.Now().Add(time.Second), Env: "dryrun", Cluster: "test"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"SLOViolationReason": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.Slodata, 0)
		if tc.params != nil && tc.params.Result == "success" {
			err = storageEsImpl.QueryCreateSloWithResult(&res, tc.params)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryCreateSloWithResult(&res, tc.params)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryCreateSloWithResult(&resErr, tc.params)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func Test_queryUpgradeSloByResult(t *testing.T) {

	type TestCase struct {
		name           string
		params         *model.SloOptions
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchUpgradeSloByResult",
			params: &model.SloOptions{Type: "create", Cluster: "clu", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Count: "5"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"UpgradeResult": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchUpgradeSloByResult",
			params: &model.SloOptions{Type: "create", Cluster: "clu", Result: "", From: time.Now(), To: time.Now().Add(time.Second), Count: "5"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"UpgradeResult": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.Slodata, 0)
		if tc.params != nil && tc.params.Result == "success" {
			res := make([]*model.Slodata, 0)
			err = storageEsImpl.QueryUpgradeSloWithResult(&res, tc.params)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryUpgradeSloWithResult(&res, tc.params)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryUpgradeSloWithResult(&resErr, tc.params)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
func Test_queryDeleteSloByResult(t *testing.T) {

	type TestCase struct {
		name           string
		params         *model.SloOptions
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchDeleteSloByResul",
			params: &model.SloOptions{Type: "create", Cluster: "cluster", Result: "success", From: time.Now(), To: time.Now().Add(time.Second), Count: "5"},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"DeleteResult": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchDeleteSloByResul",
			params: &model.SloOptions{Result: ""},
			esReqRepExpect: func() map[string]string {
				slo := &model.Slodata{}
				pb, _ := json.Marshal(slo)
				sloBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: sloBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"DeleteResult": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]*model.Slodata, 0)
		if tc.params != nil && tc.params.Result == "success" {
			err = storageEsImpl.QueryDeleteSloWithResult(&res, tc.params)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryDeleteSloWithResult(&res, tc.params)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryDeleteSloWithResult(&resErr, tc.params)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestDebuggingNodeUidParams(t *testing.T) {

	type TestCase struct {
		name           string
		params         *model.NodeParams
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchNodeWithNodename",
			params: &model.NodeParams{NodeName: "ababc"},
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				ny, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"nodeName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchNodeWithNodename",
			params: &model.NodeParams{NodeUid: "ababc"},
			esReqRepExpect: func() map[string]string {
				nodeYaml := model.NodeYaml{
					Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "abcdef"}},
				}
				ny, _ := json.Marshal(nodeYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"uid": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.NodeYaml, 0)
		if tc.params != nil {
			err = storageEsImpl.QueryNodeYamlWithParams(&res, tc.params)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryNodeYamlWithParams(&res, nil)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryNodeYamlWithParams(&resErr, tc.params)
		assert.NotEqual(t, tc.expectErr, err)
	}
}

func TestQueryAuditWithAuditId(t *testing.T) {
	type TestCase struct {
		name           string
		auditId        string
		esReqRepExpect func() map[string]string
		expectRes      model.Audit
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:    "FetchAuditWithAuditid",
			auditId: "abcdef",
			esReqRepExpect: func() map[string]string {
				audit := model.Audit{
					AuditId: "abcdef",
					AuditLog: k8saudit.Event{
						AuditID: "abcdef",
					},
				}
				pb, _ := json.Marshal(audit)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"auditID": string(b)}
			},
			expectRes: model.Audit{
				AuditId: "abcdef",
				AuditLog: k8saudit.Event{
					AuditID: "abcdef",
				},
			},
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		rs := model.Audit{}
		err = storageEsImpl.QueryAuditWithAuditId(&rs, tc.auditId)
		assert.Equal(t, tc.expectErr, err)
		assert.Equal(t, tc.expectRes, rs)
	}

}
func TestQueryEventPodsWithPodUid(t *testing.T) {
	type TestCase struct {
		name           string
		podUid         string
		esReqRepExpect func() map[string]string
		expectRes      model.Audit
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchAuditWithAuditid",
			podUid: "abcdef",
			esReqRepExpect: func() map[string]string {
				audit := model.Audit{
					AuditId: "abcdef",
					AuditLog: k8saudit.Event{
						AuditID: "abcdef",
					},
				}
				pb, _ := json.Marshal(audit)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"objectRef.uid.keyword": string(b)}
			},
			expectRes: model.Audit{
				AuditId: "abcdef",
				AuditLog: k8saudit.Event{
					AuditID: "abcdef",
				},
			},
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		rs := model.Audit{}
		err = storageEsImpl.QueryEventPodsWithPodUid(&rs, tc.podUid)
		assert.Equal(t, tc.expectErr, err)
		assert.Equal(t, tc.expectRes, rs)
	}
}
func TestQueryEventNodeWithPodUid(t *testing.T) {
	type TestCase struct {
		name           string
		podUid         string
		esReqRepExpect func() map[string]string
		expectRes      model.Audit
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchAuditWithAuditid",
			podUid: "abcdef",
			esReqRepExpect: func() map[string]string {
				audit := model.Audit{
					AuditId: "abcdef",
					AuditLog: k8saudit.Event{
						AuditID: "abcdef",
					},
				}
				pb, _ := json.Marshal(audit)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"objectRef.uid.keyword": string(b)}
			},
			expectRes: model.Audit{
				AuditId: "abcdef",
				AuditLog: k8saudit.Event{
					AuditID: "abcdef",
				},
			},
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		rs := model.Audit{}
		err = storageEsImpl.QueryEventNodeWithPodUid(&rs, tc.podUid)
		assert.Equal(t, tc.expectErr, err)
		assert.Equal(t, tc.expectRes, rs)
	}
}
func TestQueryEventWithTimeRange(t *testing.T) {
	type TestCase struct {
		name           string
		from           time.Time
		to             time.Time
		esReqRepExpect func() map[string]string
		expectRes      model.Audit
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name: "FetchAuditWithAuditid",
			from: time.Now(),
			to:   time.Now().Add(time.Second),
			esReqRepExpect: func() map[string]string {
				audit := model.Audit{
					AuditId: "abcdef",
					AuditLog: k8saudit.Event{
						AuditID: "abcdef",
					},
				}
				pb, _ := json.Marshal(audit)
				podBytes := json.RawMessage(pb)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: podBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"stageTimestamp": string(b)}
			},
			expectRes: model.Audit{
				AuditId: "abcdef",
				AuditLog: k8saudit.Event{
					AuditID: "abcdef",
				},
			},
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		rs := model.Audit{}
		err = storageEsImpl.QueryEventWithTimeRange(&rs, tc.from, tc.to)
		assert.Equal(t, tc.expectErr, err)
		assert.Equal(t, tc.expectRes, rs)
	}
}
func TestQueryPodYamlWithParams(t *testing.T) {

	type TestCase struct {
		name           string
		params         *model.PodParams
		esReqRepExpect func() map[string]string
		expectLen      int
		expectCode     int
		expectErr      error
	}
	testCases := []TestCase{
		{
			name:   "FetchNodeWithNodename",
			params: &model.PodParams{Name: "ababc"},
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				ny, _ := json.Marshal(podYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podName": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
		{
			name:   "FetchNodeWithNodename",
			params: &model.PodParams{Podip: "ababc"},
			esReqRepExpect: func() map[string]string {
				podYaml := model.PodYaml{
					PodUid: "abcdef",
					Pod:    &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "abcdef"}},
				}
				ny, _ := json.Marshal(podYaml)
				nodeBytes := json.RawMessage(ny)
				sr := elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{}, Hits: []*elastic.SearchHit{
					{Source: nodeBytes},
				}}}
				b, _ := json.Marshal(sr)
				return map[string]string{"podIP": string(b)}
			},
			expectLen:  1,
			expectCode: http.StatusOK,
			expectErr:  nil,
		},
	}

	for _, tc := range testCases {
		mock := mocks.NewMock()
		// 设置预期的值
		mock.Expects = tc.esReqRepExpect()
		client, err := mock.MockElasticSearchClient()
		// set es client
		storageEsImpl := &StorageEsImpl{
			DB: client,
		}
		res := make([]model.PodYaml, 0)
		if tc.params != nil {
			err = storageEsImpl.QueryPodYamlWithParams(&res, tc.params)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectLen, len(res))
		} else {
			err = storageEsImpl.QueryPodYamlWithParams(&res, nil)
			assert.NotEqual(t, tc.expectErr, err)
		}
		resErr := make([]string, 0)
		err = storageEsImpl.QueryPodYamlWithParams(&resErr, tc.params)
		assert.NotEqual(t, tc.expectErr, err)
	}
}
