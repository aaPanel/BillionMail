package maillog_stat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOverview_PrepareChartData(t *testing.T) {
	o := NewOverview()

	tests := []struct {
		name               string
		startTime          int64
		endTime            int64
		expectedColumnType string
	}{
		{
			name:               "less than a day -> hourly",
			startTime:          1000,
			endTime:            1000 + 3600,
			expectedColumnType: "hourly",
		},
		{
			name:               "exactly one day -> daily",
			startTime:          1000,
			endTime:            1000 + 86400,
			expectedColumnType: "daily",
		},
		{
			name:               "7 days -> daily",
			startTime:          1000,
			endTime:            1000 + 86400*7,
			expectedColumnType: "daily",
		},
		{
			name:               "30 days -> monthly",
			startTime:          1000,
			endTime:            1000 + 2592000,
			expectedColumnType: "monthly",
		},
		{
			name:               "90 days -> monthly",
			startTime:          1000,
			endTime:            1000 + 86400*90,
			expectedColumnType: "monthly",
		},
		{
			name:               "boundary: just under 1 day",
			startTime:          0,
			endTime:            86399,
			expectedColumnType: "hourly",
		},
		{
			name:               "boundary: just under 30 days",
			startTime:          0,
			endTime:            2591999,
			expectedColumnType: "daily",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columnType, xAxisField := o.prepareChartData(tt.startTime, tt.endTime)
			assert.Equal(t, tt.expectedColumnType, columnType)
			assert.NotEmpty(t, xAxisField)
		})
	}
}

func TestOverview_FillChartDataHourly(t *testing.T) {
	o := NewOverview()

	tests := []struct {
		name        string
		data        []map[string]interface{}
		fillItem    map[string]interface{}
		fillKey     string
		expectedLen int
	}{
		{
			name:        "empty data fills 24 slots",
			data:        []map[string]interface{}{},
			fillItem:    map[string]interface{}{"sends": 0},
			fillKey:     "x",
			expectedLen: 24,
		},
		{
			name: "partial data fills gaps",
			data: []map[string]interface{}{
				{"x": 0, "sends": 10},
				{"x": 12, "sends": 20},
			},
			fillItem:    map[string]interface{}{"sends": 0},
			fillKey:     "x",
			expectedLen: 24,
		},
		{
			name: "full 24 hours returned as-is",
			data: func() []map[string]interface{} {
				d := make([]map[string]interface{}, 24)
				for i := 0; i < 24; i++ {
					d[i] = map[string]interface{}{"x": i, "sends": i * 10}
				}
				return d
			}(),
			fillItem:    map[string]interface{}{"sends": 0},
			fillKey:     "x",
			expectedLen: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.fillChartDataHourly(tt.data, tt.fillItem, tt.fillKey)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestOverview_FillChartDataDaily(t *testing.T) {
	o := NewOverview()

	tests := []struct {
		name      string
		data      []map[string]interface{}
		fillItem  map[string]interface{}
		fillKey   string
		startTime int64
		endTime   int64
		minLen    int
	}{
		{
			name:      "negative startTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: -1,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "negative endTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 100,
			endTime:   -1,
			minLen:    1,
		},
		{
			name:      "startTime > endTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 200,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "3-day range creates 3 entries",
			data:      []map[string]interface{}{},
			fillItem:  map[string]interface{}{"sends": 0},
			fillKey:   "x",
			startTime: 1700000000,
			endTime:   1700000000 + 86400*2,
			minLen:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.fillChartDataDaily(tt.data, tt.fillItem, tt.fillKey, tt.startTime, tt.endTime)
			assert.GreaterOrEqual(t, len(result), tt.minLen)
		})
	}
}

func TestOverview_FillChartDataMonthly(t *testing.T) {
	o := NewOverview()

	tests := []struct {
		name      string
		data      []map[string]interface{}
		fillItem  map[string]interface{}
		fillKey   string
		startTime int64
		endTime   int64
		minLen    int
	}{
		{
			name:      "negative startTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: -1,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "negative endTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 100,
			endTime:   -1,
			minLen:    1,
		},
		{
			name:      "startTime > endTime returns data as-is",
			data:      []map[string]interface{}{{"x": "2024-01", "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 200,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "3-month range",
			data:      []map[string]interface{}{},
			fillItem:  map[string]interface{}{"sends": 0},
			fillKey:   "x",
			startTime: 1700000000,
			endTime:   1700000000 + 2592000*2,
			minLen:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.fillChartDataMonthly(tt.data, tt.fillItem, tt.fillKey, tt.startTime, tt.endTime)
			assert.GreaterOrEqual(t, len(result), tt.minLen)
		})
	}
}

func TestOverview_FillChartData_Dispatch(t *testing.T) {
	o := NewOverview()
	fillItem := map[string]interface{}{"val": 0}

	t.Run("hourly dispatch", func(t *testing.T) {
		result := o.fillChartData(nil, fillItem, "hourly", "x", 0, 0)
		assert.Len(t, result, 24)
	})

	t.Run("daily dispatch", func(t *testing.T) {
		result := o.fillChartData(nil, fillItem, "daily", "x", 1700000000, 1700000000+86400*2)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("monthly dispatch", func(t *testing.T) {
		result := o.fillChartData(nil, fillItem, "monthly", "x", 1700000000, 1700000000+2592000*2)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("unknown type returns data as-is", func(t *testing.T) {
		data := []map[string]interface{}{{"x": 1}}
		result := o.fillChartData(data, fillItem, "unknown", "x", 0, 0)
		assert.Equal(t, data, result)
	})
}

func TestOverview_FilterAndPrepareTimeSection(t *testing.T) {
	o := NewOverview()

	t.Run("both positive unchanged", func(t *testing.T) {
		s, e := o.filterAndPrepareTimeSection(100, 200)
		assert.Equal(t, int64(100), s)
		assert.Equal(t, int64(200), e)
	})

	t.Run("startTime positive endTime negative -> endTime set to now", func(t *testing.T) {
		now := time.Now().Unix()
		startTime := now - 3600 // 1 hour ago, within 1-year max range
		s, e := o.filterAndPrepareTimeSection(startTime, -1)
		assert.Equal(t, startTime, s)
		assert.InDelta(t, now, e, 2)
	})

	t.Run("max 1 year range enforced", func(t *testing.T) {
		end := int64(1700000000)
		start := end - 50000000 // ~578 days, more than 1 year
		s, e := o.filterAndPrepareTimeSection(start, end)
		assert.Equal(t, end, e)
		assert.Equal(t, end-31622400, s, "start should be clamped to 1 year before end")
	})

	t.Run("exactly 1 year range passes", func(t *testing.T) {
		end := int64(1700000000)
		start := end - 31622400
		s, e := o.filterAndPrepareTimeSection(start, end)
		assert.Equal(t, start, s)
		assert.Equal(t, end, e)
	})

	t.Run("panic on startTime > endTime", func(t *testing.T) {
		assert.Panics(t, func() {
			o.filterAndPrepareTimeSection(200, 100)
		})
	})
}

func TestOverview_FillChartDataHourly_Idempotent(t *testing.T) {
	o := NewOverview()
	fillItem := map[string]interface{}{"val": 0}

	data := []map[string]interface{}{
		{"x": 5, "val": 42},
	}

	// First call
	result1 := o.fillChartDataHourly(data, fillItem, "x")
	assert.Len(t, result1, 24)

	// The original data item should still be at position 5
	assert.Equal(t, 42, result1[5]["val"])
}

func TestOverview_FillChartDataDaily_PreservesExisting(t *testing.T) {
	o := NewOverview()

	// Use a known date
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local).Unix()
	fillItem := map[string]interface{}{"sends": 0}
	data := []map[string]interface{}{
		{"x": "2024-01-15", "sends": 100},
	}

	result := o.fillChartDataDaily(data, fillItem, "x", day1, day1+86400)
	assert.GreaterOrEqual(t, len(result), 1)

	// Find the entry for our date
	found := false
	for _, item := range result {
		if item["sends"] == 100 {
			found = true
			break
		}
	}
	assert.True(t, found, "original data should be preserved")
}
