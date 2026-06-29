package load

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Metrics 采集压测指标与时间序列。
type Metrics struct {
	mu             sync.Mutex
	bucketInterval time.Duration
	latencies      []time.Duration
	total          int
	success        int
	failed         int
	errors         map[string]int
	buckets        map[int]*bucket
	stepStats      map[string]*stepAccumulator
	startedAt      time.Time
	endedAt        time.Time
}

// NewMetrics 创建指标采集器。
func NewMetrics(bucketInterval time.Duration) *Metrics {
	if bucketInterval <= 0 {
		bucketInterval = time.Second
	}
	return &Metrics{
		bucketInterval: bucketInterval,
		errors:         make(map[string]int),
		buckets:        make(map[int]*bucket),
	}
}

// MarkStarted 记录压测开始时间（仅首次有效）。
func (m *Metrics) MarkStarted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.startedAt.IsZero() {
		m.startedAt = time.Now()
	}
}

// MarkEnded 记录压测结束时间。
func (m *Metrics) MarkEnded() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endedAt = time.Now()
}

// Record 记录一次 scenario 执行结果。
func (m *Metrics) Record(latency time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureStarted()
	m.total++
	if err != nil {
		m.failed++
		key := classifyError(err)
		m.errors[key]++
	} else {
		m.success++
		m.latencies = append(m.latencies, latency)
	}

	offset := m.bucketIndexLocked(time.Now())
	m.bucketLocked(offset).add(latency, err)
}

// RecordCustom 记录 scenario 上报的业务指标（按当前时间桶聚合）。
func (m *Metrics) RecordCustom(name string, value float64) {
	if name == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureStarted()
	offset := m.bucketIndexLocked(time.Now())
	m.bucketLocked(offset).addCustom(name, value)
}

// Snapshot 返回当前指标快照。
func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	latencies := append([]time.Duration(nil), m.latencies...)
	errors := make(map[string]int, len(m.errors))
	for k, v := range m.errors {
		errors[k] = v
	}

	snap := Snapshot{
		Total:          m.total,
		Success:        m.success,
		Failed:         m.failed,
		Errors:         errors,
		StartedAt:      m.startedAt,
		EndedAt:        m.endedAt,
		BucketInterval: m.bucketInterval,
		Steps:          m.stepSummariesLocked(),
	}
	snap.computeLatency(latencies)
	snap.buildTimeSeries(m.buckets, m.bucketInterval)
	return snap
}

func (m *Metrics) ensureStarted() {
	if m.startedAt.IsZero() {
		m.startedAt = time.Now()
	}
}

func (m *Metrics) bucketIndexLocked(at time.Time) int {
	elapsed := at.Sub(m.startedAt)
	if elapsed < 0 {
		return 0
	}
	return int(elapsed / m.bucketInterval)
}

func (m *Metrics) bucketLocked(index int) *bucket {
	if m.buckets[index] == nil {
		m.buckets[index] = newBucket()
	}
	return m.buckets[index]
}

// Snapshot 指标快照，用于报告序列化。
type Snapshot struct {
	Total          int                      `json:"total"`
	Success        int                      `json:"success"`
	Failed         int                      `json:"failed"`
	SuccessPct     float64                  `json:"success_pct"`
	ActualQPS      float64                  `json:"actual_qps"`
	Errors         map[string]int           `json:"errors"`
	Latency        LatencySummary           `json:"latency"`
	Steps          []StepSummary            `json:"steps,omitempty"`
	TimeSeries     []TimePoint              `json:"time_series"`
	CustomSeries   map[string][]CustomPoint `json:"custom_series,omitempty"`
	StartedAt      time.Time                `json:"started_at"`
	EndedAt        time.Time                `json:"ended_at"`
	BucketInterval time.Duration            `json:"bucket_interval"`
}

// LatencySummary 延迟统计。
type LatencySummary struct {
	Min time.Duration `json:"min"`
	Avg time.Duration `json:"avg"`
	P50 time.Duration `json:"p50"`
	P90 time.Duration `json:"p90"`
	P95 time.Duration `json:"p95"`
	P99 time.Duration `json:"p99"`
	Max time.Duration `json:"max"`
}

func (s *Snapshot) buildTimeSeries(buckets map[int]*bucket, interval time.Duration) {
	if len(buckets) == 0 {
		return
	}

	indices := make([]int, 0, len(buckets))
	for idx := range buckets {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	customNames := map[string]struct{}{}
	for _, b := range buckets {
		for name := range b.custom {
			customNames[name] = struct{}{}
		}
	}

	s.TimeSeries = make([]TimePoint, 0, len(indices))
	s.CustomSeries = make(map[string][]CustomPoint)

	for _, idx := range indices {
		offsetSec := int(float64(idx) * interval.Seconds())
		tp := buckets[idx].toTimePoint(offsetSec, interval)
		s.TimeSeries = append(s.TimeSeries, tp)

		for name, cp := range buckets[idx].toCustomPoints(offsetSec) {
			s.CustomSeries[name] = append(s.CustomSeries[name], cp)
		}
	}
}

func (s *Snapshot) computeLatency(latencies []time.Duration) {
	if len(latencies) == 0 {
		if s.Total > 0 {
			s.SuccessPct = float64(s.Success) / float64(s.Total) * 100
		}
		return
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var sum time.Duration
	for _, d := range latencies {
		sum += d
	}

	s.Latency = LatencySummary{
		Min: latencies[0],
		Avg: sum / time.Duration(len(latencies)),
		P50: percentileOfSorted(latencies, 0.50),
		P90: percentileOfSorted(latencies, 0.90),
		P95: percentileOfSorted(latencies, 0.95),
		P99: percentileOfSorted(latencies, 0.99),
		Max: latencies[len(latencies)-1],
	}

	if s.Total > 0 {
		s.SuccessPct = float64(s.Success) / float64(s.Total) * 100
	}
	elapsed := s.EndedAt.Sub(s.StartedAt).Seconds()
	if elapsed > 0 {
		s.ActualQPS = float64(s.Total) / elapsed
	}
}

// percentileOfSorted 从已升序排列的延迟样本中取分位数。
func percentileOfSorted(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func classifyError(err error) string {
	if err == nil {
		return "ok"
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "context deadline exceeded"), strings.Contains(msg, "timeout"):
		return "timeout"
	case strings.Contains(msg, "unexpected status"):
		return "http_error"
	case strings.Contains(msg, "unexpected response"):
		return "business_error"
	default:
		return "error"
	}
}
