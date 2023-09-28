package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/queue"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apiserver/pkg/apis/audit"
)

const rs2Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandString generate n-length random string
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = rs2Letters[rand.Intn(len(rs2Letters))]
	}
	return string(b)
}

func TimeSinceInMilliSeconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Millisecond.Nanoseconds())
}

func TimeSinceInSeconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Second.Nanoseconds())
}

// TimeSinceInMinutes return minutes since start
func TimeSinceInMinutes(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Minute.Nanoseconds())
}

func TimeDiffInMilliSeconds(start, end time.Time) float64 {
	if start.IsZero() || end.IsZero() {
		return -1
	}
	return float64(end.Sub(start).Nanoseconds() / time.Millisecond.Nanoseconds())
}

// TimeDiffInSeconds get time difference in unit second.
func TimeDiffInSeconds(start, end time.Time) float64 {
	if start.IsZero() || end.IsZero() {
		return -1
	}
	return float64(end.Sub(start).Nanoseconds() / time.Second.Nanoseconds())
}

// TimeDiffInMilliMinutes return time difference in minutes
func TimeDiffInMilliMinutes(start, end time.Time) float64 {
	if start.IsZero() || end.IsZero() {
		return -1
	}
	return float64(end.Sub(start).Nanoseconds() / time.Minute.Nanoseconds())
}

// Marshal Marshal object to string of the error message if failed
func Marshal(v interface{}) string {
	bs, e := json.Marshal(v)
	if e != nil {
		return e.Error()
	}
	return string(bs)
}

// Dumps dump interface as indented json object
func Dumps(v interface{}) string {
	bs, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		klog.Errorf("MarshalError(%s): %+v", err.Error(), v)
		return fmt.Sprintf("%+v", v)
	}
	return string(bs)
}

// DumpsEventBasic will remove some info from pod spec to reduce log size and sensitive data to be leaked
func DumpsEventBasic(event *audit.Event) string {
	if event.ObjectRef.Resource == "pods" && (event.Verb == "create" || event.Verb == "update") {
		if event.RequestObject == nil || event.RequestObject.Raw == nil {
			return Dumps(event)
		}
		tmp := event.RequestObject.Raw
		pod := corev1.Pod{}
		if err := json.Unmarshal(tmp, &pod); err == nil {
			pod.Spec = corev1.PodSpec{}
		}
		if bs, err := json.Marshal(pod); err == nil {
			event.RequestObject.Raw = bs
		}
	}
	return Dumps(event)
}

func Warnf(format string, args ...interface{}) {
	f := "\n####################################\n####################################\n\n"
	f += format
	f += "\n\n####################################\n####################################\n"
	klog.Warningf(f, args...)
}

func DumpsEventKeyInfo(event *audit.Event) string {
	s := fmt.Sprintf("%s: %s: %s at %+v", event.AuditID, event.Verb, event.RequestURI, event.StageTimestamp.Time.Format("2006/01/02 15:04:05.000000"))
	return s
}

// ReTry try to call a function until success.
func ReTry(f func() error, interval time.Duration, retryCount int) (err error) {
	needContinue := true
	i := 0

	for needContinue {
		i++
		err = f()
		if err != nil && i <= retryCount {
			needContinue = true
		} else {
			needContinue = false
		}
		if needContinue {
			time.Sleep(interval)
		}
	}

	return err
}

// GetPullingImageFromEventMessage get image name from pulling event
func GetPullingImageFromEventMessage(s string) string {
	re := regexp.MustCompile(`([pP]ulling image ")(\w.*)(")`)
	sm := re.FindStringSubmatch(s)
	if len(sm) == 4 {
		return validImage(sm[2])
	}
	return ""
}

// GetPulledImageFromEventMessage get image name from pulled event
func GetPulledImageFromEventMessage(s string) string {
	re := regexp.MustCompile(`(Successfully pulled image ")(\w.*)(")`)
	sm := re.FindStringSubmatch(s)
	if len(sm) == 4 {
		return validImage(sm[2])
	}

	re = regexp.MustCompile(`(Container image ")(\w.*)(" already present on machine)`)
	sm = re.FindStringSubmatch(s)
	if len(sm) == 4 {
		return validImage(sm[2])
	}

	return ""
}

func validImage(image string) string {
	// FIXME valid image, but really needed ?
	return image
}

func MergeMaps(mapTo, mapFrom map[string]string) {
	if mapFrom == nil {
		return
	}

	for k, v := range mapFrom {
		mapTo[k] = v
	}
}

// CloneStringMap copy src to dst of type map[string]string
func CloneStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string)
	if src == nil {
		return dst
	}

	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// ShouldIgnoreNamespace return true is namespace should be omit
func ShouldIgnoreNamespace(ns string) bool {
	nses := strings.Split(os.Getenv("IGNORED_NAMESPACES"), ",")
	for i := range nses {
		if nses[i] == ns {
			return true
		}
	}

	return false
}

func IgnorePanic(desc string) {
	if err := recover(); err != nil {
		klog.Error(desc, err)
		logPanic(err)
	}
}

func StringHashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func SliceContainsString(arr []string, s string) bool {
	if s == "" {
		return false
	}
	for _, ss := range arr {
		if ss == s {
			return true
		}
	}
	return false
}

type BufferData struct {
	data       map[string]interface{}
	parent     *BufferData
	sequential bool
	sync.WaitGroup
}

func (d *BufferData) waitParent() {
	if d.parent != nil {
		d.parent.Wait()
	}
}

func (d *BufferData) setParent(parent *BufferData) {
	if d.sequential {
		d.parent = parent
	}
}

func (d *BufferData) clear() {
	d.data = nil
}

func NewBufferData(sequential bool) *BufferData {
	bd := &BufferData{
		data:       make(map[string]interface{}),
		WaitGroup:  sync.WaitGroup{},
		sequential: sequential,
	}
	bd.Add(1)

	return bd
}

type BufferUtils struct {
	name        string
	bufferSize  int
	bufferData  *BufferData
	saveFunc    func(map[string]interface{}) error
	clearPeriod time.Duration
	workQueue   *queue.BoundedQueue
	sequential  bool
	sync.Mutex
}

func NewBufferUtils(name string, bufferSize int, clearPeriod time.Duration, sequential bool, saveFunc func(map[string]interface{}) error) *BufferUtils {
	bufferUtils := &BufferUtils{
		name:        name,
		bufferSize:  bufferSize,
		clearPeriod: clearPeriod,
		bufferData:  NewBufferData(sequential),
		sequential:  sequential,
		saveFunc:    saveFunc,
	}
	queue := queue.NewBoundedQueue(fmt.Sprintf("buffer-util-%s", name), 1000, nil)
	queue.StartLengthReporting(10 * time.Second)
	queue.IsDropEventOnFull = false

	bufferUtils.workQueue = queue
	return bufferUtils
}

func (b *BufferUtils) SaveData(id string, data interface{}) error {
	var readyData *BufferData = nil
	b.Lock()
	b.bufferData.data[id] = data
	if len(b.bufferData.data) > b.bufferSize {
		readyData = b.bufferData
		b.bufferData = NewBufferData(b.sequential)
		b.bufferData.setParent(readyData)
	}
	b.Unlock()

	if b.saveFunc == nil {
		return nil
	}

	_ = b.workQueue.Produce(readyData)
	return nil
}

func (b *BufferUtils) DoClearData() {
	stopChan := make(<-chan struct{})
	go wait.Until(func() {
		b.Lock()
		readyData := b.bufferData
		b.bufferData = NewBufferData(b.sequential)
		b.bufferData.setParent(readyData)
		b.Unlock()

		klog.V(8).Infof("do clear data for %s, size: %d", b.name, len(readyData.data))
		b.workQueue.Produce(readyData)
	}, b.clearPeriod, stopChan)

	b.workQueue.StartConsumers(20, func(v interface{}) {
		defer IgnorePanic("StartConsumers_BufferUtils ")

		bufferData, ok := v.(*BufferData)
		if !ok || bufferData == nil {
			return
		}
		bufferData.waitParent()
		err := b.saveFunc(bufferData.data)
		if err != nil {
			klog.Error("Do clear data error", err)
		}
		bufferData.Done()
		bufferData.clear()
	})
}

// wait to clear data when get stop signal
func (b *BufferUtils) Stop() {
	klog.V(8).Infof("do clear when stop for %s, data size: %d", b.name, len(b.bufferData.data))
	b.workQueue.Produce(b.bufferData)
	for b.workQueue.Size() > 0 {
		time.Sleep(100 * time.Millisecond)
	}

	b.workQueue.Stop()
}

func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func HandleCrash() {
	if r := recover(); r != nil {
		logPanic(r)
	}
}

func logPanic(r interface{}) {
	callers := getCallers(r)
	klog.Errorf("Observed a panic: %#v (%v)\n%v", r, r, callers)
}

func getCallers(r interface{}) string {
	callers := ""
	for i := 0; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	return callers
}

// 计算 Pod 的 CPU 和 mem
func CalculateCpuAndMem(pod *corev1.Pod) (float64, float64) {
	var podCpuMilli float64 = 0
	var podMemMilli float64 = 0
	for _, container := range pod.Spec.Containers {
		podCpuMilli += float64(container.Resources.Requests.Cpu().MilliValue())
		podMemMilli += float64(container.Resources.Requests.Memory().MilliValue())
	}
	return podCpuMilli / float64(1000), podMemMilli / float64(1000*1024*1024)
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func SetStringParam(values url.Values, name string, f *string) {
	s := values.Get(name)
	if s != "" {
		*f = s
	}
}

func SetBoolParam(values url.Values, name string, f *bool) {
	s := values.Get(name)
	if s == "t" || s == "true" || s == "1" {
		*f = true
	}
}

func SetTimeParam(values url.Values, name string, f *time.Time) {
	s := values.Get(name)
	if s != "" {
		t, err := ParseTime(s)
		if err != nil {
			klog.Errorf("failed to parse time %s for %s: %s", s, name, err.Error())
		} else {
			*f = t
		}
	}
}

func SetTimeLayoutParam(values url.Values, name string, f *time.Time) {
	s := values.Get(name)
	layOut := "2006-01-02T15:04:05"
	if s != "" {
		t, err := time.ParseInLocation(layOut, s, time.Local)
		if err != nil {
			klog.Errorf("failed to parse time %s for %s: %s", s, name, err.Error())
		} else {
			*f = t
		}
	}
}

// trace param parse

func ParseTags(simpleTags []string, jsonTags []string) (map[string]string, error) {
	retMe := make(map[string]string)
	for _, tag := range simpleTags {
		keyAndValue := strings.Split(tag, ":")
		if l := len(keyAndValue); l > 1 {
			retMe[keyAndValue[0]] = strings.Join(keyAndValue[1:], ":")
		} else {
			return nil, fmt.Errorf("malformed 'tag' parameter, expecting key:value, received: %s", tag)
		}
	}
	for _, tags := range jsonTags {
		var fromJSON map[string]string
		if err := json.Unmarshal([]byte(tags), &fromJSON); err != nil {
			return nil, fmt.Errorf("malformed 'tags' parameter, cannot unmarshal JSON: %s", err)
		}
		for k, v := range fromJSON {
			retMe[k] = v
		}
	}
	return retMe, nil
}

type Tabler interface {
	TableName() string
}

type EsTabler interface {
	EsTableName() string
}

type Typer interface {
	TypeName() string
}

// 获取数据的 SqlTableName EsTableName TypeName
func GetMetaName(dest interface{}) (string, string, string, error) {
	value := reflect.ValueOf(dest)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}
	modelType := reflect.Indirect(value).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return "", "", "", errors.New("data must be struct or slice or array")
	}
	modelValue := reflect.New(modelType)
	tabler, ok := modelValue.Interface().(Tabler)
	if !ok {
		return "", "", "", errors.New("data doesn't have TableName")
	}
	esTabler, ok := modelValue.Interface().(EsTabler)
	if !ok {
		return "", "", "", errors.New("data doesn't have TableName")
	}
	typer, ok := modelValue.Interface().(Typer)
	if !ok {
		return "", "", "", errors.New("data doesn't have TableName")
	}
	return tabler.TableName(), esTabler.EsTableName(), typer.TypeName(), nil
}
