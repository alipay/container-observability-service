package data_access

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSqlQuerySpanWithPodUid(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podUid    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QuerySpanWithPodUid",
			podUid:    "abcdef",
			res:       make([]*model.Span, 0),
			expectErr: nil,
		},
		{
			name:      "QuerySpanWithPodUid",
			podUid:    "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QuerySpanWithPodUid",
			podUid:    "",
			res:       make([]*model.Span, 0),
			expectErr: nil,
		},
	}
	// uid := "aaa"
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"doc_id", "or_resource", "or_namespace", "or_name", "or_uid", "or_apigroup"}).
		AddRow("123", "", "", "", "123", "")

	mock.ExpectQuery("SELECT * FROM `span` WHERE or_uid = ? ORDER BY span_elapsed desc LIMIT 80").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	// res := make([]*model.Span, 0)
	for _, tt := range testCases {
		if tt.podUid == "abcdef" {
			err = fClient.QuerySpanWithPodUid(&tt.res, tt.podUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podUid == "123" || tt.podUid == "" {
			err = fClient.QuerySpanWithPodUid(&tt.res, tt.podUid)
			assert.NotEqual(t, tt.expectErr, err)
		}

	}
}
func TestSqlQueryLifePhaseWithPodUid(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podUid    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QuerySpanWithPodUid",
			podUid:    "abcdef",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryLifePhaseWithPodUid",
			podUid:    "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryLifePhaseWithPodUid",
			podUid:    "",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"doc_id", "cluster_name", "namespace", "podName", "pod_uid", "operation_name"}).
		AddRow("123", "", "", "", "123", "")

	mock.ExpectQuery("SELECT * FROM `pod_phase` WHERE pod_uid = ? ORDER BY start_time desc LIMIT 200").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.podUid == "abcdef" {
			err = fClient.QueryLifePhaseWithPodUid(&tt.res, tt.podUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podUid == "123" || tt.podUid == "" {
			err = fClient.QueryLifePhaseWithPodUid(&tt.res, tt.podUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlsWithPodUID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podUid    string
		res       []*model.PodYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodYamlsWithPodUID",
			podUid:    "abcdef",
			res:       []*model.PodYaml{{ClusterName: "cluster", Namespace: "space"}},
			expectErr: nil,
		},
		{
			name:      "QueryPodYamlsWithPodUID",
			podUid:    "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodYamlsWithPodUID",
			podUid:    "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname", "pod"}).
		AddRow("123", "", "", "", "123", "", "{\"kind\":\"Pod\", \"apiVersion\":\"v1\"}")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_uid = ? ORDER BY stage_timestamp desc LIMIT 100").WithArgs("abcdef").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podUid == "abcdef" {
			err = fClient.QueryPodYamlsWithPodUID(&tt.res, tt.podUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podUid == "123" || tt.podUid == "" {
			err = fClient.QueryPodYamlsWithPodUID(tt.res, tt.podUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlsWithPodName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podName   string
		res       []*model.PodYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QuerySpanWithPodUid",
			podName:   "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QuerySpanWithPodUid",
			podName:   "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodYamlsWithPodUID",
			podName:   "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname", "pod"}).
		AddRow("123", "", "", "", "123", "", "{\"kind\":\"Pod\", \"apiVersion\":\"v1\"}")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_name = ? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podName == "abcdef" {
			err = fClient.QueryPodYamlsWithPodName(&tt.res, tt.podName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podName == "123" || tt.podName == "" {
			err = fClient.QueryPodYamlsWithPodName(&tt.res, tt.podName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlsWithHostName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		hostName  string
		res       []*model.PodYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "PodYamlsWithHostName",
			hostName:  "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "PodYamlsWithHostName",
			hostName:  "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodYamlsWithPodUID",
			hostName:  "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname", "pod"}).
		AddRow("123", "", "", "", "123", "", "{\"kind\":\"Pod\", \"apiVersion\":\"v1\"}")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE hostname =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.hostName == "abcdef" {
			err = fClient.QueryPodYamlsWithHostName(&tt.res, tt.hostName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.hostName == "123" || tt.hostName == "" {
			err = fClient.QueryPodYamlsWithHostName(&tt.res, tt.hostName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlsWithPodIp(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podIp     string
		res       []*model.PodYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "PodYamlsWithHostName",
			podIp:     "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "PodYamlsWithHostName",
			podIp:     "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodYamlsWithPodUID",
			podIp:     "123",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname", "pod"}).
		AddRow("123", "", "", "", "123", "", "{\"kind\":\"Pod\", \"apiVersion\":\"v1\"}")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_ip = ? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podIp == "abcdef" {
			err = fClient.QueryPodYamlsWithPodIp(&tt.res, tt.podIp)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podIp == "123" || tt.podIp == "" {
			err = fClient.QueryPodYamlsWithPodIp(&tt.res, tt.podIp)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodListWithNodeip(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeIp    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname", "is_deleted"}).
		AddRow("123", "", "", "", "123", "123", false)

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE host_ip =?  AND is_deleted =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.nodeIp == "abcdef" {
			err = fClient.QueryPodListWithNodeip(&tt.res, tt.nodeIp, false)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeIp == "123" || tt.nodeIp == "" {
			err = fClient.QueryPodListWithNodeip(&tt.res, tt.nodeIp, false)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodUIDListByHostname(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		hostName  string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			hostName:  "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			hostName:  "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			hostName:  "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname"}).
		AddRow("123", "", "", "", "123", "123")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE hostname =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.hostName == "abcdef" {
			err = fClient.QueryPodUIDListByHostname(&tt.res, tt.hostName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.hostName == "123" || tt.hostName == "" {
			err = fClient.QueryPodUIDListByHostname(&tt.res, tt.hostName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodUIDListByPodIP(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podIp     string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			podIp:     "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podIp:     "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podIp:     "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname"}).
		AddRow("123", "", "", "1234", "123", "123")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_ip =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podIp == "abcdef" {
			err = fClient.QueryPodUIDListByPodIP(&tt.res, tt.podIp)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podIp == "123" || tt.podIp == "" {
			err = fClient.QueryPodUIDListByPodIP(&tt.res, tt.podIp)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodUIDListByPodName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podName   string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			podName:   "abcdef",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podName:   "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podName:   "",
			res:       make([]*model.PodYaml, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_name", "pod_uid", "hostname"}).
		AddRow("123", "", "", "abcde", "123", "123")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_name =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.podName == "abcdef" {
			err = fClient.QueryPodUIDListByPodName(&tt.res, tt.podName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podName == "123" || tt.podName == "" {
			err = fClient.QueryPodUIDListByPodName(&tt.res, tt.podName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryNodeYamlsWithNodeUid(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		NodeUid   string
		res       []*model.NodeYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryNodeYamlsWithNodeUid",
			NodeUid:   "abcdef",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryNodeYamlsWithNodeUid",
			NodeUid:   "",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "node", "node_name", "node_ip", "uid"}).
		AddRow("123", "", "{\"kind\":\"Node\", \"apiVersion\":\"v1\"}", "abcde", "123", "123")

	mock.ExpectQuery("SELECT * FROM `node_yaml` WHERE uid =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.NodeUid == "abcdef" {
			err = fClient.QueryNodeYamlsWithNodeUid(&tt.res, tt.NodeUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.NodeUid == "123" || tt.NodeUid == "" {
			err = fClient.QueryNodeYamlsWithNodeUid(&tt.res, tt.NodeUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryNodeYamlsWithNodeName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeName  string
		res       []*model.NodeYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "NodeYamlsWithNodeName",
			nodeName:  "abcdef",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},

		{
			name:      "NodeYamlsWithNodeName",
			nodeName:  "",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "node", "node_name", "node_ip", "uid"}).
		AddRow("123", "", "{\"kind\":\"Node\", \"apiVersion\":\"v1\"}", "abcde", "123", "123")

	mock.ExpectQuery("SELECT * FROM `node_yaml` WHERE node_name =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.nodeName == "abcdef" {
			err = fClient.QueryNodeYamlsWithNodeName(&tt.res, tt.nodeName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeName == "123" || tt.nodeName == "" {
			err = fClient.QueryNodeYamlsWithNodeName(&tt.res, tt.nodeName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryNodeYamlsWithNodeIP(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeIp    string
		res       []*model.NodeYaml
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "NodeYamlsWithnodeIp",
			nodeIp:    "abcdef",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
		{
			name:      "NodeYamlsWithnodeIp",
			nodeIp:    "",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "node", "node_name", "node_ip", "uid"}).
		AddRow("123", "", "{\"kind\":\"Node\", \"apiVersion\":\"v1\"}", "abcde", "123", "123")

	mock.ExpectQuery("SELECT * FROM `node_yaml` WHERE node_ip =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.nodeIp == "abcdef" {
			err = fClient.QueryNodeYamlsWithNodeIP(&tt.res, tt.nodeIp)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeIp == "123" || tt.nodeIp == "" {
			err = fClient.QueryNodeYamlsWithNodeIP(&tt.res, tt.nodeIp)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryNodeUIDListWithNodeIp(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeIp    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "abcdef",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeIp:    "",
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "node", "node_name", "node_ip", "uid"}).
		AddRow("123", "", "", "abcde", "123", "123")

	mock.ExpectQuery("SELECT * FROM `node_yaml` WHERE node_ip =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.nodeIp == "abcdef" {
			err = fClient.QueryNodeUIDListWithNodeIp(&tt.res, tt.nodeIp)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeIp == "123" || tt.nodeIp == "" {
			err = fClient.QueryNodeUIDListWithNodeIp(&tt.res, tt.nodeIp)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodInfoWithPodUid(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podUid    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "abcdef",
			res:       make([]*model.PodInfo, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "",
			res:       make([]*model.PodInfo, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"doc_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "pod_name"}).
		AddRow("123", "", "", "1234", "123", "123")

	mock.ExpectQuery("SELECT * FROM `slo_pod_info` WHERE pod_uid =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podUid == "abcdef" {
			err = fClient.QueryPodInfoWithPodUid(&tt.res, tt.podUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podUid == "123" || tt.podUid == "" {
			err = fClient.QueryPodInfoWithPodUid(&tt.res, tt.podUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlsWithNodeIP(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	nodeIp := "abcd"
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "host_ip", "pod_uid", "hostname"}).
		AddRow("123", "", "", "abcd", "123", "123")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE host_ip =? AND is_deleted =? ORDER BY stage_timestamp desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	res := make([]*model.PodYaml, 0)

	err = fClient.QueryPodYamlsWithNodeIP(&res, nodeIp)
	assert.Nil(t, err)
}

func TestSqlQueryNodephaseWithNodeUID(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeUid   string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			nodeUid:   "abcdef",
			res:       make([]*model.NodeLifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeUid:   "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeUid:   "",
			res:       make([]*model.NodeLifePhase, 0),
			expectErr: nil,
		},
	}
	rows := sqlmock.NewRows([]string{"doc_id", "cluster_name", "node_name", "node_uid", "operation_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa")

	mock.ExpectQuery("SELECT * FROM `node_phase` WHERE node_uid = ? ORDER BY start_time desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.nodeUid == "abcdef" {
			err = fClient.QueryNodephaseWithNodeUID(&tt.res, tt.nodeUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeUid == "123" || tt.nodeUid == "" {
			err = fClient.QueryNodephaseWithNodeUID(&tt.res, tt.nodeUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryNodephaseWithNodeName(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		nodeName  string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			nodeName:  "abcdef",
			res:       make([]*model.NodeLifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeName:  "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			nodeName:  "",
			res:       make([]*model.NodeLifePhase, 0),
			expectErr: nil,
		},
	}

	rows := sqlmock.NewRows([]string{"doc_id", "cluster_name", "node_name", "node_uid", "operation_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa")

	mock.ExpectQuery("SELECT * FROM `node_phase` WHERE node_name = ? ORDER BY start_time desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.nodeName == "abcdef" {
			err = fClient.QueryNodephaseWithNodeName(&tt.res, tt.nodeName)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.nodeName == "123" || tt.nodeName == "" {
			err = fClient.QueryNodephaseWithNodeName(&tt.res, tt.nodeName)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryDebuggingWithPodUid(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		podUid    string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "abcdef",
			res:       make([]*model.SloTraceData, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			podUid:    "",
			res:       make([]*model.SloTraceData, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"doc_id", "cluster", "namespace", "pod_ip", "pod_uid", "node_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa", "dsaaa")

	mock.ExpectQuery("SELECT * FROM `slo_trace_data_daily` WHERE pod_uid =? ORDER BY created_time desc LIMIT 100").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.podUid == "abcdef" {
			err = fClient.QuerySloTraceDataWithPodUID(&tt.res, tt.podUid)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.podUid == "123" || tt.podUid == "" {
			err = fClient.QuerySloTraceDataWithPodUID(&tt.res, tt.podUid)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryCreateSloWithResult(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		params    model.SloOptions
		res       interface{}
		expectErr error
		querySql  string
	}
	rows := sqlmock.NewRows([]string{"doc_id", "cluster", "namespace", "pod_name", "pod_uid", "node_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa", "dsaaa")
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{BizName: "12345", Result: "success", Type: "create"},
			res:       make([]*model.Slodata, 0),
			expectErr: nil,
			querySql:  "SELECT * FROM `slo_data` WHERE slo_violation_reason=? AND biz_name =? ORDER BY created desc LIMIT 300",
		},
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{Cluster: "cluster", Result: "success", Type: "create"},
			res:       make([]*model.Slodata, 0),
			expectErr: nil,
			querySql:  "SELECT * FROM `slo_data` WHERE slo_violation_reason=? AND cluster =? ORDER BY created desc LIMIT 300",
		},
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{From: time.Now(), Result: "success", Type: "create"},
			res:       make([]*model.Slodata, 0),
			expectErr: nil,
			querySql:  "SELECT * FROM `slo_data` WHERE slo_violation_reason=? AND created >=? ORDER BY created desc LIMIT 300",
		},
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{SloTime: "10s", Result: "success", Type: "create"},
			res:       make([]*model.Slodata, 0),
			expectErr: nil,
			querySql:  "SELECT * FROM `slo_data` WHERE slo_violation_reason=? AND pod_slo=? ORDER BY created desc LIMIT 300",
		},
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{DeliveryStatus: "success", Result: "success", Type: "create"},
			res:       make([]*model.Slodata, 0),
			expectErr: nil,
			querySql:  "SELECT * FROM `slo_data` WHERE slo_violation_reason=? AND delivery_status_orig =? ORDER BY created desc LIMIT 300",
		},
	}
	// 构造模拟的查询结果集

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		mock.ExpectQuery(tt.querySql).WillReturnRows(rows)
		err = fClient.QueryCreateSloWithResult(&tt.res, &tt.params)
		assert.Equal(t, tt.expectErr, err)
	}
}
func TestSqlQueryUpgradeSloWithResult(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		params    model.SloOptions
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{BizName: "12345", Result: "success", Type: "create"},
			res:       make([]*model.SloTraceData, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			params:    model.SloOptions{Cluster: "abc", Result: "success", Type: "create"},
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"doc_id", "cluster", "namespace", "pod_name", "pod_uid", "node_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa", "dsaaa")

	mock.ExpectQuery("SELECT * FROM `slo_trace_data_daily` WHERE upgrade_result =? AND type =? ORDER BY created_time desc LIMIT 300").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.params.BizName == "12345" {
			err = fClient.QueryUpgradeSloWithResult(&tt.res, &tt.params)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.params.Cluster == "abc" {
			err = fClient.QueryUpgradeSloWithResult(&tt.res, &tt.params)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryDeleteSloWithResult(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		params    model.SloOptions
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryDeleteSloWithResult",
			params:    model.SloOptions{BizName: "12345", Result: "success", Type: "create"},
			res:       make([]*model.SloTraceData, 0),
			expectErr: nil,
		},
		{
			name:      "QueryDeleteSloWithResult",
			params:    model.SloOptions{Cluster: "abc", Result: "success", Type: "create"},
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"doc_id", "cluster", "namespace", "pod_name", "pod_uid", "node_name"}).
		AddRow("dsdsa", "", "", "", "dsdaa", "dsaaa")

	mock.ExpectQuery("SELECT * FROM `slo_trace_data_daily` WHERE delete_result =? AND type= ? ORDER BY created_time desc LIMIT 300").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.params.BizName == "12345" {
			err = fClient.QueryDeleteSloWithResult(&tt.res, &tt.params)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.params.Cluster == "abc" {
			err = fClient.QueryDeleteSloWithResult(&tt.res, &tt.params)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlDebuggingNodeUidParams(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		params    model.NodeParams
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryDeleteSloWithResult",
			params:    model.NodeParams{NodeUid: "12345", NodeName: "success", NodeIp: "create"},
			res:       make([]*model.NodeYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "node_name", "node_ip", "uid", "cluster_name", "node"}).
		AddRow("dsdsa", "", "", "", "dsdaa", "dsaaa")

	mock.ExpectQuery("SELECT * FROM `node_yaml` WHERE node_name =? ORDER BY stage_timestamp desc LIMIT 1").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.params.NodeUid == "12345" {
			err = fClient.QueryNodeYamlWithParams(&tt.res, &tt.params)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.params.NodeUid == "abc" {
			err = fClient.QueryNodeYamlWithParams(&tt.res, &tt.params)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryAuditWithAuditId(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		auditId   string
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryPodListWithNodeip",
			auditId:   "abcdef",
			res:       make([]*model.Audit, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			auditId:   "123",
			res:       make([]*model.LifePhase, 0),
			expectErr: nil,
		},
		{
			name:      "QueryPodListWithNodeip",
			auditId:   "",
			res:       make([]*model.Audit, 0),
			expectErr: nil,
		},
	}

	rows := sqlmock.NewRows([]string{"audit_id", "cluster", "namespace", "resource", "content"}).
		AddRow("dsdsa", "", "", "", "dsdaa")

	mock.ExpectQuery("SELECT * FROM `audit` WHERE audit_id = ?").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)

	for _, tt := range testCases {
		if tt.auditId == "abcdef" {
			err = fClient.QueryAuditWithAuditId(&tt.res, tt.auditId)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.auditId == "123" || tt.auditId == "" {
			err = fClient.QueryAuditWithAuditId(&tt.res, tt.auditId)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
func TestSqlQueryPodYamlWithParams(t *testing.T) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	type TestCase struct {
		name      string
		params    model.PodParams
		res       interface{}
		expectErr error
	}
	testCases := []TestCase{
		{
			name:      "QueryDeleteSloWithResult",
			params:    model.PodParams{Uid: "12345", Name: "success", Podip: "create"},
			res:       make([]model.PodYaml, 0),
			expectErr: nil,
		},
	}
	// 构造模拟的查询结果集
	rows := sqlmock.NewRows([]string{"audit_id", "cluster_name", "namespace", "pod_ip", "pod_uid", "hostname"}).
		AddRow("123", "", "", "", "123", "")

	mock.ExpectQuery("SELECT * FROM `pod_yaml` WHERE pod_name =? ORDER BY stage_timestamp desc LIMIT 10").WillReturnRows(rows)

	fClient := &StorageSqlImpl{
		DB: gormDB,
	}
	gostub.Stub(&XSearch, fClient)
	for _, tt := range testCases {
		if tt.params.Uid == "12345" {
			err = fClient.QueryPodYamlWithParams(&tt.res, &tt.params)
			assert.Equal(t, tt.expectErr, err)
		}
		if tt.params.Uid == "abc" {
			err = fClient.QueryPodYamlWithParams(&tt.res, &tt.params)
			assert.NotEqual(t, tt.expectErr, err)
		}
	}
}
