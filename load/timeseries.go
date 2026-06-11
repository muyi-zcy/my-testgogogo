package load

import "sort"

import "time"

// TimePoint 单个时间桶的框架指标。
type TimePoint struct {
	OffsetSec  int           `json:"offset_sec"`
	Requests   int           `json:"requests"`
	Success    int           `json:"success"`
	Failed     int           `json:"failed"`
	QPS        float64       `json:"qps"`
	SuccessPct float64       `json:"success_pct"`
	LatencyP95 time.Duration `json:"latency_p95"`
}

// CustomPoint 单个时间桶的自定义指标聚合。
type CustomPoint struct {
	OffsetSec int     `json:"offset_sec"`
	Avg       float64 `json:"avg"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Count     int     `json:"count"`
}

// bucket 内部时间桶累积器。
type bucket struct {
	requests  int
	success   int
	failed    int
	latencies []time.Duration
	custom    map[string][]float64
}

func newBucket() *bucket {
	return &bucket{custom: make(map[string][]float64)}
}

func (b *bucket) add(latency time.Duration, err error) {
	b.requests++
	if err != nil {
		b.failed++
		return
	}
	b.success++
	b.latencies = append(b.latencies, latency)
}

func (b *bucket) addCustom(name string, value float64) {
	b.custom[name] = append(b.custom[name], value)
}

func (b *bucket) toTimePoint(offsetSec int, interval time.Duration) TimePoint {
	tp := TimePoint{
		OffsetSec: offsetSec,
		Requests:  b.requests,
		Success:   b.success,
		Failed:    b.failed,
	}
	if b.requests > 0 {
		tp.SuccessPct = float64(b.success) / float64(b.requests) * 100
	}
	if interval > 0 {
		tp.QPS = float64(b.requests) / interval.Seconds()
	}
	if len(b.latencies) > 0 {
		sorted := append([]time.Duration(nil), b.latencies...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
		tp.LatencyP95 = percentile(sorted, 0.95)
	}
	return tp
}

func (b *bucket) toCustomPoints(offsetSec int) map[string]CustomPoint {
	out := make(map[string]CustomPoint, len(b.custom))
	for name, values := range b.custom {
		if len(values) == 0 {
			continue
		}
		min, max := values[0], values[0]
		var sum float64
		for _, v := range values {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		out[name] = CustomPoint{
			OffsetSec: offsetSec,
			Avg:       sum / float64(len(values)),
			Min:       min,
			Max:       max,
			Count:     len(values),
		}
	}
	return out
}
