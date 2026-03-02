package agent

import (
	"maps"
	"slices"
	"testing"
	"time"

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
	return New("", 1, 1, true, nil, 1, 1, zap.L())
}

func TestRetryTimeouts(t *testing.T) {
	testCases := []struct {
		name     string
		timeouts []time.Duration
		interval time.Duration
		expected []time.Duration
	}{
		{
			name:     "empty timeouts slice",
			timeouts: []time.Duration{},
			interval: 1 * time.Second,
			expected: []time.Duration{},
		},
		{
			name:     "single timeout less than interval",
			timeouts: []time.Duration{1 * time.Second},
			interval: 5 * time.Second,
			expected: []time.Duration{1 * time.Second},
		},
		{
			name:     "single timeout equal to interval",
			timeouts: []time.Duration{1 * time.Second},
			interval: 1 * time.Second,
			expected: []time.Duration{},
		},
		{
			name:     "single timeout greater than interval",
			timeouts: []time.Duration{5 * time.Second},
			interval: 1 * time.Second,
			expected: []time.Duration{},
		},
		{
			name:     "multiple timeouts with sum less than interval",
			timeouts: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			interval: 10 * time.Second,
			expected: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
		},
		{
			name:     "multiple timeouts with sum equal to interval",
			timeouts: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			interval: 6 * time.Second,
			expected: []time.Duration{1 * time.Second, 2 * time.Second},
		},
		{
			name:     "multiple timeouts with sum greater than interval",
			timeouts: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			interval: 3 * time.Second,
			expected: []time.Duration{1 * time.Second},
		},
		{
			name:     "multiple timeouts with sum greater than interval (second case)",
			timeouts: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			interval: 4 * time.Second,
			expected: []time.Duration{1 * time.Second, 2 * time.Second},
		},
		{
			name:     "interval zero",
			timeouts: []time.Duration{1 * time.Second, 2 * time.Second},
			interval: 0 * time.Second,
			expected: []time.Duration{},
		},
		{
			name:     "large timeouts",
			timeouts: []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second},
			interval: 15 * time.Second,
			expected: []time.Duration{10 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := retryTimeouts(tc.timeouts, tc.interval)
			assert.Equal(t, tc.expected, result)
		})
	}
}
