package utils

// 流量控制，核心思路：
// 	采用快失败、慢恢复的策略

type FlowControl struct {
	window               []bool  // 最近WindowSize次查询记录
	WindowMaxSize        int     // 窗口大小
	FastBackoffThreshold int     // 快速退避错误阈值
	FastBackoffFactor    float64 // 快速退避因子
	SlowRecoverFactor    float64 // 慢恢复因子
	MaxFlowSize          float64 // 最大流量值
	latestFailCount      int     // 最近连续失败次数
	lastFlowSize         float64 // 最近一次流量大小
}

func NewFlowControl(wms int, fbt int, fbf float64, srf float64, mfs float64) *FlowControl {
	fc := &FlowControl{
		window:               make([]bool, wms, wms),
		WindowMaxSize:        wms,
		FastBackoffThreshold: fbt,
		FastBackoffFactor:    fbf,
		SlowRecoverFactor:    srf,
		MaxFlowSize:          mfs,
		lastFlowSize:         mfs,
	}

	if fc.WindowMaxSize < 5 {
		fc.WindowMaxSize = 5
	}

	if fc.FastBackoffThreshold > fc.WindowMaxSize {
		fc.FastBackoffThreshold = fc.WindowMaxSize
	}

	if fc.SlowRecoverFactor < 1 {
		fc.SlowRecoverFactor = 1.2
	}

	return fc
}

func (f *FlowControl) RecordFlow(flowSuccess bool) {
	if f.WindowMaxSize == 0 {
		return
	}

	if len(f.window)+1 > f.WindowMaxSize {
		f.window = f.window[1:]
	}
	f.window = append(f.window, flowSuccess)
	if flowSuccess {
		f.latestFailCount = 0
	} else {
		f.latestFailCount += 1
	}

}

func (f *FlowControl) DecideFlow(desire float64) float64 {
	if f.latestFailCount >= f.FastBackoffThreshold {
		f.lastFlowSize = f.lastFlowSize * f.FastBackoffFactor
		if f.lastFlowSize < 1.0 {
			f.lastFlowSize = 1.0
		}
		return f.lastFlowSize
	}

	// just use last size, when last error

	if len(f.window) > 0 && !f.window[len(f.window)-1] {
		return f.lastFlowSize
	}

	// decide new flow
	slowRecoverTarget := f.lastFlowSize * f.SlowRecoverFactor

	if slowRecoverTarget > desire {
		slowRecoverTarget = desire
	}

	if slowRecoverTarget > f.MaxFlowSize {
		f.lastFlowSize = f.MaxFlowSize
	} else {
		f.lastFlowSize = slowRecoverTarget
	}

	return f.lastFlowSize
}
