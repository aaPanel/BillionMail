package batch_mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTaskTagIds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []int{},
		},
		{
			name:     "single tag",
			input:    "[1]",
			expected: []int{1},
		},
		{
			name:     "multiple tags",
			input:    "[1,2,3]",
			expected: []int{1, 2, 3},
		},
		{
			name:     "large tag IDs",
			input:    "[100,200,999]",
			expected: []int{100, 200, 999},
		},
		{
			name:     "bare number not parseable as array",
			input:    "5",
			expected: []int{}, // gconv.Scan can't unmarshal a bare int into []int
		},
		{
			name:     "invalid JSON returns nil-coerced-to-empty",
			input:    "not-valid",
			expected: []int(nil), // gconv.Scan fails, returns the zero-value nil slice
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: []int{},
		},
		{
			name:     "spaces in JSON array",
			input:    "[1, 2, 3]",
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTaskTagIds(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContactFilter_Struct(t *testing.T) {
	// Test that ContactFilter can be created with expected field values
	filter := ContactFilter{
		GroupId:  10,
		TagIds:   []int{1, 2, 3},
		TagLogic: "AND",
	}
	assert.Equal(t, 10, filter.GroupId)
	assert.Equal(t, []int{1, 2, 3}, filter.TagIds)
	assert.Equal(t, "AND", filter.TagLogic)

	// OR logic
	filter2 := ContactFilter{
		GroupId:  0,
		TagIds:   []int{5},
		TagLogic: "OR",
	}
	assert.Equal(t, 0, filter2.GroupId)
	assert.Equal(t, "OR", filter2.TagLogic)

	// NOT logic
	filter3 := ContactFilter{
		TagLogic: "NOT",
		TagIds:   []int{7, 8},
	}
	assert.Equal(t, "NOT", filter3.TagLogic)
}

func TestCreateTaskArgs_Struct(t *testing.T) {
	args := CreateTaskArgs{
		Addresser:   "sender@example.com",
		Subject:     "Test Subject",
		FullName:    "Test User",
		TemplateId:  1,
		IsRecord:    1,
		Unsubscribe: 1,
		Threads:     4,
		TrackOpen:   1,
		TrackClick:  1,
		Remark:      "test remark",
		StartTime:   1700000000,
		Warmup:      1,
		AddType:     0,
		GroupId:     5,
		TagIds:      []int{1, 2},
		TagLogic:    "AND",
	}

	assert.Equal(t, "sender@example.com", args.Addresser)
	assert.Equal(t, "Test Subject", args.Subject)
	assert.Equal(t, 4, args.Threads)
	assert.Equal(t, []int{1, 2}, args.TagIds)
	assert.Equal(t, "AND", args.TagLogic)
}
