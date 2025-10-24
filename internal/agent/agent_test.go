package agent

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCollect(t *testing.T) {
	a := testAgent(t)
	a.Collect()

	expectedGauges := []string{
		MetricAlloc,
		MetricBuckHashSys,
		MetricFrees,
		MetricGCCPUFraction,
		MetricGCSys,
		MetricHeapAlloc,
		MetricHeapIdle,
		MetricHeapInuse,
		MetricHeapObjects,
		MetricHeapReleased,
		MetricHeapSys,
		MetricLastGC,
		MetricLookups,
		MetricMCacheInuse,
		MetricMCacheSys,
		MetricMSpanInuse,
		MetricMSpanSys,
		MetricMallocs,
		MetricNextGC,
		MetricNumForcedGC,
		MetricNumGC,
		MetricOtherSys,
		MetricPauseTotalNs,
		MetricStackInuse,
		MetricStackSys,
		MetricSys,
		MetricTotalAlloc,
		MetricRandomValue,
	}

	expectedCounters := []string{
		MetricPollCount,
	}

	gauges := slices.Collect(maps.Keys(a.gauges))
	counters := slices.Collect(maps.Keys(a.counters))

	for _, k := range expectedGauges {
		assert.Contains(t, gauges, k)
	}
	for _, k := range expectedCounters {
		assert.Contains(t, counters, k)
	}
}

func testAgent(t *testing.T) *Agent {
	t.Helper()
	return New("", 1, 1, 100, zap.L())
}
