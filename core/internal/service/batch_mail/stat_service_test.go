package batch_mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareChartData(t *testing.T) {
	svc := NewTaskStatService()

	tests := []struct {
		name               string
		startTime          int64
		endTime            int64
		expectedColumnType string
	}{
		{
			name:               "less than a day -> hourly",
			startTime:          1000,
			endTime:            1000 + 3600, // 1 hour
			expectedColumnType: "hourly",
		},
		{
			name:               "exactly one day -> daily",
			startTime:          1000,
			endTime:            1000 + 86400,
			expectedColumnType: "daily",
		},
		{
			name:               "multi-day -> daily",
			startTime:          1000,
			endTime:            1000 + 86400*7,
			expectedColumnType: "daily",
		},
		{
			name:               "half day -> hourly",
			startTime:          0,
			endTime:            43200,
			expectedColumnType: "hourly",
		},
		{
			name:               "30 days -> daily",
			startTime:          0,
			endTime:            86400 * 30,
			expectedColumnType: "daily",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columnType, xAxisField := svc.prepareChartData(tt.startTime, tt.endTime)
			assert.Equal(t, tt.expectedColumnType, columnType)
			assert.NotEmpty(t, xAxisField)
		})
	}
}

func TestFillChartDataHourly(t *testing.T) {
	svc := NewTaskStatService()

	tests := []struct {
		name     string
		data     []map[string]interface{}
		fillItem map[string]interface{}
		fillKey  string
		expected int // expected length
	}{
		{
			name:     "empty data -> 24 hours filled",
			data:     []map[string]interface{}{},
			fillItem: map[string]interface{}{"sends": 0},
			fillKey:  "x",
			expected: 24,
		},
		{
			name: "partial data -> gaps filled",
			data: []map[string]interface{}{
				{"x": 0, "sends": 10},
				{"x": 12, "sends": 20},
				{"x": 23, "sends": 5},
			},
			fillItem: map[string]interface{}{"sends": 0},
			fillKey:  "x",
			expected: 24,
		},
		{
			name: "int64 type values",
			data: []map[string]interface{}{
				{"x": int64(5), "sends": 15},
			},
			fillItem: map[string]interface{}{"sends": 0},
			fillKey:  "x",
			expected: 24,
		},
		{
			name: "float64 type values",
			data: []map[string]interface{}{
				{"x": float64(10), "sends": 30},
			},
			fillItem: map[string]interface{}{"sends": 0},
			fillKey:  "x",
			expected: 24,
		},
		{
			name: "string type values",
			data: []map[string]interface{}{
				{"x": "15", "sends": 7},
			},
			fillItem: map[string]interface{}{"sends": 0},
			fillKey:  "x",
			expected: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.fillChartDataHourly(tt.data, tt.fillItem, tt.fillKey)
			assert.Len(t, result, tt.expected)

			// Verify all hours 0-23 are present
			for i := 0; i < 24; i++ {
				assert.NotNil(t, result[i])
				assert.Contains(t, result[i], tt.fillKey)
			}
		})
	}
}

func TestFillChartDataHourly_PreservesExistingData(t *testing.T) {
	svc := NewTaskStatService()
	data := []map[string]interface{}{
		{"x": 5, "sends": 42, "delivered": 40},
		{"x": 10, "sends": 100, "delivered": 95},
	}
	fillItem := map[string]interface{}{"sends": 0, "delivered": 0}

	result := svc.fillChartDataHourly(data, fillItem, "x")

	assert.Equal(t, 42, result[5]["sends"])
	assert.Equal(t, 40, result[5]["delivered"])
	assert.Equal(t, 100, result[10]["sends"])
	assert.Equal(t, 95, result[10]["delivered"])

	// Filled slots should have default values
	assert.Equal(t, 0, result[0]["sends"])
	assert.Equal(t, 0, result[0]["delivered"])
}

func TestFillChartDataDaily(t *testing.T) {
	svc := NewTaskStatService()

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
			data:      []map[string]interface{}{{"x": int64(100), "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: -1,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "negative endTime returns data as-is",
			data:      []map[string]interface{}{{"x": int64(100), "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 100,
			endTime:   -1,
			minLen:    1,
		},
		{
			name:      "startTime > endTime returns data as-is",
			data:      []map[string]interface{}{{"x": int64(200), "y": 1}},
			fillItem:  map[string]interface{}{"y": 0},
			fillKey:   "x",
			startTime: 200,
			endTime:   100,
			minLen:    1,
		},
		{
			name:      "3-day range fills gaps",
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
			result := svc.fillChartDataDaily(tt.data, tt.fillItem, tt.fillKey, tt.startTime, tt.endTime)
			assert.GreaterOrEqual(t, len(result), tt.minLen)
		})
	}
}

func TestFillChartData_Dispatch(t *testing.T) {
	svc := NewTaskStatService()
	fillItem := map[string]interface{}{"val": 0}

	t.Run("hourly dispatch", func(t *testing.T) {
		result := svc.fillChartData(nil, fillItem, "hourly", "x", 0, 0)
		assert.Len(t, result, 24)
	})

	t.Run("daily dispatch", func(t *testing.T) {
		result := svc.fillChartData(nil, fillItem, "daily", "x", 1700000000, 1700000000+86400*2)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("unknown type returns data as-is", func(t *testing.T) {
		data := []map[string]interface{}{{"x": 1}}
		result := svc.fillChartData(data, fillItem, "unknown", "x", 0, 0)
		assert.Equal(t, data, result)
	})
}
