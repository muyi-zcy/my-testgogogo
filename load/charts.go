package load

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// renderTimeSeriesSection 渲染压测过程表格与 Mermaid 图表。
func renderTimeSeriesSection(m Snapshot, opts ReportOptions) string {
	if len(m.TimeSeries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## 压测过程\n\n")

	b.WriteString("| 时间(s) | 请求 | QPS | 成功率 | p95 | 失败 |\n")
	b.WriteString("|---------|------|-----|--------|-----|------|\n")
	for _, tp := range m.TimeSeries {
		p95 := "-"
		if tp.LatencyP95 > 0 {
			p95 = tp.LatencyP95.Round(time.Millisecond).String()
		}
		b.WriteString(fmt.Sprintf("| %d | %d | %.1f | %.1f%% | %s | %d |\n",
			tp.OffsetSec, tp.Requests, tp.QPS, tp.SuccessPct, p95, tp.Failed))
	}
	b.WriteString("\n")

	if !opts.Charts {
		return b.String()
	}

	labels := timeSeriesLabels(m.TimeSeries)
	qpsValues := make([]float64, len(m.TimeSeries))
	p95Values := make([]float64, len(m.TimeSeries))
	successValues := make([]float64, len(m.TimeSeries))
	for i, tp := range m.TimeSeries {
		qpsValues[i] = tp.QPS
		p95Values[i] = float64(tp.LatencyP95.Milliseconds())
		successValues[i] = tp.SuccessPct
	}

	b.WriteString("### QPS\n\n")
	b.WriteString(renderMermaidXYChart("QPS", "req/s", labels, qpsValues))
	b.WriteString("\n")

	if hasPositive(p95Values) {
		b.WriteString("### 延迟 p95\n\n")
		b.WriteString(renderMermaidXYChart("Latency p95", "ms", labels, p95Values))
		b.WriteString("\n")
	}

	b.WriteString("### 成功率\n\n")
	b.WriteString(renderMermaidXYChart("Success rate", "%", labels, successValues))
	b.WriteString("\n")

	chartMetrics := append([]string(nil), opts.ChartMetrics...)
	if len(chartMetrics) == 0 {
		for name := range m.CustomSeries {
			chartMetrics = append(chartMetrics, name)
		}
	}
	sort.Strings(chartMetrics)

	for _, name := range chartMetrics {
		series, ok := m.CustomSeries[name]
		if !ok || len(series) == 0 {
			continue
		}
		cLabels := make([]string, len(series))
		values := make([]float64, len(series))
		for i, cp := range series {
			cLabels[i] = fmt.Sprintf("%d", cp.OffsetSec)
			values[i] = cp.Avg
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", name))
		b.WriteString(renderMermaidXYChart(name, "avg", cLabels, values))
		b.WriteString("\n")
	}

	return b.String()
}

func timeSeriesLabels(points []TimePoint) []string {
	labels := make([]string, len(points))
	for i, tp := range points {
		labels[i] = fmt.Sprintf("%d", tp.OffsetSec)
	}
	return labels
}

func renderMermaidXYChart(title, yLabel string, xLabels []string, values []float64) string {
	if len(xLabels) == 0 || len(values) == 0 {
		return ""
	}

	yMax := maxValue(values)
	if yMax <= 0 {
		yMax = 1
	}
	yMax = math.Ceil(yMax * 1.1)

	var b strings.Builder
	b.WriteString("```mermaid\n")
	b.WriteString("xychart-beta\n")
	b.WriteString(fmt.Sprintf("    title \"%s\"\n", escapeMermaid(title)))
	b.WriteString("    x-axis [")
	b.WriteString(strings.Join(xLabels, ", "))
	b.WriteString("]\n")
	b.WriteString(fmt.Sprintf("    y-axis \"%s\" 0 --> %.0f\n", escapeMermaid(yLabel), yMax))
	b.WriteString("    line [")
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%.1f", v)
	}
	b.WriteString(strings.Join(parts, ", "))
	b.WriteString("]\n")
	b.WriteString("```\n")
	return b.String()
}

func maxValue(values []float64) float64 {
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func hasPositive(values []float64) bool {
	for _, v := range values {
		if v > 0 {
			return true
		}
	}
	return false
}

func escapeMermaid(s string) string {
	return strings.ReplaceAll(s, "\"", "'")
}
