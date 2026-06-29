package load

import (
	"sort"
	"time"
)

type stepAccumulator struct {
	total   int
	success int
	failed  int
	latency []time.Duration
}

// StepSummary 单个 Flow 步骤的压测统计。
type StepSummary struct {
	Name    string         `json:"name"`
	Total   int            `json:"total"`
	Success int            `json:"success"`
	Failed  int            `json:"failed"`
	Latency LatencySummary `json:"latency"`
}

// RecordStep 记录 Flow scenario 中某一步的执行结果。
func (m *Metrics) RecordStep(name string, latency time.Duration, err error) {
	if name == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureStarted()
	if m.stepStats == nil {
		m.stepStats = make(map[string]*stepAccumulator)
	}
	acc := m.stepStats[name]
	if acc == nil {
		acc = &stepAccumulator{}
		m.stepStats[name] = acc
	}
	acc.total++
	if err != nil {
		acc.failed++
	} else {
		acc.success++
		acc.latency = append(acc.latency, latency)
	}
}

func (m *Metrics) stepSummariesLocked() []StepSummary {
	if len(m.stepStats) == 0 {
		return nil
	}
	names := make([]string, 0, len(m.stepStats))
	for name := range m.stepStats {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]StepSummary, 0, len(names))
	for _, name := range names {
		acc := m.stepStats[name]
		summary := StepSummary{
			Name:    name,
			Total:   acc.total,
			Success: acc.success,
			Failed:  acc.failed,
		}
		summary.Latency = summarizeLatencies(acc.latency)
		out = append(out, summary)
	}
	return out
}

func summarizeLatencies(latencies []time.Duration) LatencySummary {
	if len(latencies) == 0 {
		return LatencySummary{}
	}
	sorted := append([]time.Duration(nil), latencies...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}
	return LatencySummary{
		Min: sorted[0],
		Avg: sum / time.Duration(len(sorted)),
		P50: percentileOfSorted(sorted, 0.50),
		P90: percentileOfSorted(sorted, 0.90),
		P95: percentileOfSorted(sorted, 0.95),
		P99: percentileOfSorted(sorted, 0.99),
		Max: sorted[len(sorted)-1],
	}
}
