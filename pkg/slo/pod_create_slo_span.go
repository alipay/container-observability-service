package slo

import (
	"sort"
	"time"

	"github.com/alipay/container-observability-service/pkg/spans"
	"k8s.io/apimachinery/pkg/types"
)

type SpanSlice []spans.Span

func (s SpanSlice) Len() int           { return len(s) }
func (s SpanSlice) Less(i, j int) bool { return s[i].Begin.Before(s[j].Begin) }
func (s SpanSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SpanSlicePtr []*spans.Span

func (s SpanSlicePtr) Len() int           { return len(s) }
func (s SpanSlicePtr) Less(i, j int) bool { return s[i].Begin.Before(s[j].Begin) }
func (s SpanSlicePtr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func ReorgSpans(podSpans SpanSlice) (open, customopen bool, lastclose, firstopenspanbeg time.Time, dspans SpanSlice) {
	// sort span by span.Begin
	sort.Sort(podSpans)

	open = false
	customopen = false
	zeroTime := time.Unix(0, 0)
	lastclose = zeroTime
	firstopenspanbeg = zeroTime

	dspans = SpanSlice{}
	for _, s := range podSpans {
		if s.Begin.IsZero() {
			continue
		}

		// skip others span
		if s.GetConfig() == nil || (s.GetConfig().SpanOwner == spans.OthersOwner || s.GetConfig().SpanOwner == "") {
			continue
		}

		// custom span
		if s.GetConfig().SpanOwner == spans.CustomOwner {
			if s.End.IsZero() {
				customopen = true
			} else {
				if s.End.After(lastclose) {
					lastclose = s.End
				}
			}
			continue
		}

		if s.GetConfig().SpanOwner == spans.K8sOwner {
			if s.End.IsZero() {
				open = true
				if firstopenspanbeg == zeroTime {
					firstopenspanbeg = s.Begin
				}

				return
			} else {
				dspans = append(dspans, s)

				if s.End.After(lastclose) {
					lastclose = s.End
				}
			}
		}
	}

	return
}

func ReorgSpansNew(podSpans SpanSlicePtr, currentTime time.Time) (SpanSlicePtr, SpanSlicePtr, time.Duration) {
	// sort span by span.Begin
	sort.Sort(podSpans)

	k8sSpans := make([]*spans.Span, 0, podSpans.Len())
	customSpans := make([]*spans.Span, 0, podSpans.Len())

	var totalCost time.Duration = 0
	for _, s := range podSpans {
		if s.Begin.IsZero() {
			continue
		}

		// find left and right
		if totalCost == 0 {
			totalCost = currentTime.Sub(s.Begin)
		}

		// skip others span
		if s.GetConfig() == nil || (s.GetConfig().SpanOwner == spans.OthersOwner || s.GetConfig().SpanOwner == "") {
			continue
		}

		// add end time for open span
		if s.End.IsZero() {
			s.End = currentTime
			s.Elapsed = s.End.Sub(s.Begin).Milliseconds()
		}

		// custom span
		if s.GetConfig().SpanOwner == spans.CustomOwner {
			customSpans = append(customSpans, s)
		}

		if s.GetConfig().SpanOwner == spans.K8sOwner {
			k8sSpans = append(k8sSpans, s)
		}

	}

	return customSpans, k8sSpans, totalCost
}

func DeliveryTimeCalc(podSpans SpanSlice, currentTime time.Time) time.Duration {
	if len(podSpans) == 0 {
		return 0
	}

	open, customopen, lastclose, firstopenspanbeg, dspans := ReorgSpans(podSpans)

	closecost := greedyCal(dspans)

	if open {
		return closecost + currentTime.Sub(firstopenspanbeg)
	}

	if customopen {
		return closecost
	}

	zeroTime := time.Unix(0, 0)
	var wc time.Duration = 0
	if lastclose != zeroTime {
		wc = currentTime.Sub(lastclose)
	}

	return closecost + wc
}

func DeliveryTimeCalcNew(podSpans SpanSlicePtr, currentTime time.Time, uid types.UID) time.Duration {
	if len(podSpans) == 0 {
		return 0
	}
	customSpans, k8sSpans, totalCost := ReorgSpansNew(podSpans, currentTime)

	customCost := calculateCustomCost(customSpans, k8sSpans, uid)

	return totalCost - customCost
}

func greedyCal(spans SpanSlice) time.Duration {
	var dur time.Duration = 0

	idx := 0
	maxTime := time.Unix(^int64(0), 0)
	zeroTime := time.Unix(0, 0)
	start := zeroTime
	end := zeroTime

	for idx < len(spans) {
		span := spans[idx]
		idx++

		if !end.After(span.Begin) {
			dur += end.Sub(start)
			start = zeroTime
			end = maxTime
		}

		if start == zeroTime {
			start = span.Begin
			end = span.End
		}

		if span.End.After(end) {
			end = span.End
		}
	}

	dur += end.Sub(start)

	return dur
}

func calculateCustomCost(customSpans, k8sSpans SpanSlicePtr, uid types.UID) time.Duration {
	if customSpans == nil || len(customSpans) == 0 {
		return 0
	}

	//防止遍历为空
	if k8sSpans == nil {
		k8sSpans = SpanSlicePtr{}
	}

	currentCustomSPans := customSpans
	for _, ks := range k8sSpans {
		nextCustomSPans := SpanSlicePtr{}
		for _, cs := range currentCustomSPans {
			if cs.Elapsed == 0 {
				continue
			}

			if !ks.End.After(cs.Begin) || !ks.Begin.Before(cs.End) {
				nextCustomSPans = append(nextCustomSPans, cs)
				continue
			}

			//case one
			if !ks.Begin.After(cs.Begin) && !ks.End.Before(cs.End) {
				continue
			}
			//case two
			if ks.Begin.After(cs.Begin) && ks.End.Before(cs.End) {
				left := spans.Span{Type: cs.Type, Name: cs.Name, Begin: cs.Begin, End: ks.Begin}
				right := spans.Span{Type: cs.Type, Name: cs.Name, Begin: ks.End, End: cs.End}
				left.Elapsed = left.End.Sub(left.Begin).Milliseconds()
				right.Elapsed = right.End.Sub(right.Begin).Milliseconds()

				nextCustomSPans = append(nextCustomSPans, &left)
				nextCustomSPans = append(nextCustomSPans, &right)
				continue
			}
			//case three
			if ks.End.After(cs.Begin) && !ks.Begin.After(cs.Begin) {
				newSpan := spans.Span{Type: cs.Type, Name: cs.Name, Begin: ks.End, End: cs.End}
				newSpan.Elapsed = newSpan.End.Sub(newSpan.Begin).Milliseconds()
				nextCustomSPans = append(nextCustomSPans, &newSpan)
				continue
			}
			//case foure
			if ks.Begin.After(cs.Begin) && !ks.End.Before(cs.End) {
				newSpan := spans.Span{Type: cs.Type, Name: cs.Name, Begin: cs.Begin, End: ks.Begin}
				newSpan.Elapsed = newSpan.End.Sub(newSpan.Begin).Milliseconds()
				nextCustomSPans = append(nextCustomSPans, &newSpan)
			}
		}

		currentCustomSPans = nextCustomSPans
	}

	// sort span by span.Begin
	sort.Sort(currentCustomSPans)
	// greedy compact
	cost := time.Duration(0)
	var begin, end time.Time

	for _, cs := range currentCustomSPans {
		if begin.IsZero() || end.IsZero() {
			begin = cs.Begin
			end = cs.End
		}

		if cs.Begin.After(end) {
			cost += end.Sub(begin)
			begin = cs.Begin
			end = cs.End
		} else if cs.End.After(end) {
			end = cs.End
		}
	}

	cost += end.Sub(begin)
	return cost
}
