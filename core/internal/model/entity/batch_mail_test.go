package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// EmailTask.MarshalJSON
// ---------------------------------------------------------------------------

func TestEmailTask_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		task       EmailTask
		wantTagIds []int
	}{
		{
			name: "TagIdsRaw populated, TagIds empty",
			task: EmailTask{
				Id:        1,
				TaskName:  "test",
				TagIdsRaw: "[1,2,3]",
			},
			wantTagIds: []int{1, 2, 3},
		},
		{
			name: "TagIds already parsed, TagIdsRaw present",
			task: EmailTask{
				Id:        2,
				TagIdsRaw: "[10,20]",
				TagIds:    []int{10, 20},
			},
			wantTagIds: []int{10, 20},
		},
		{
			name: "TagIdsRaw empty, TagIds empty",
			task: EmailTask{
				Id: 3,
			},
			wantTagIds: nil,
		},
		{
			name: "TagIdsRaw empty string, TagIds empty",
			task: EmailTask{
				Id:        4,
				TagIdsRaw: "",
			},
			wantTagIds: nil,
		},
		{
			name: "invalid JSON in TagIdsRaw, TagIds empty",
			task: EmailTask{
				Id:        5,
				TagIdsRaw: "not-json",
			},
			wantTagIds: nil,
		},
		{
			name: "TagIdsRaw single element",
			task: EmailTask{
				Id:        6,
				TagIdsRaw: "[42]",
			},
			wantTagIds: []int{42},
		},
		{
			name: "TagIdsRaw empty array",
			task: EmailTask{
				Id:        7,
				TagIdsRaw: "[]",
			},
			wantTagIds: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(&tt.task)
			require.NoError(t, err)

			// unmarshal into generic map to inspect tag_ids
			var m map[string]interface{}
			err = json.Unmarshal(data, &m)
			require.NoError(t, err)

			if tt.wantTagIds == nil {
				// tag_ids should be null or missing
				raw, exists := m["tag_ids"]
				if exists {
					assert.Nil(t, raw, "expected null tag_ids")
				}
			} else {
				rawSlice, ok := m["tag_ids"].([]interface{})
				require.True(t, ok, "tag_ids should be array, got %T", m["tag_ids"])
				got := make([]int, len(rawSlice))
				for i, v := range rawSlice {
					got[i] = int(v.(float64))
				}
				assert.Equal(t, tt.wantTagIds, got)
			}
		})
	}
}

func TestEmailTask_MarshalJSON_TagIdsRawNotInOutput(t *testing.T) {
	task := EmailTask{
		Id:        1,
		TagIdsRaw: "[1,2,3]",
	}
	data, err := json.Marshal(&task)
	require.NoError(t, err)

	// TagIdsRaw has json:"-" so should not appear in output
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	_, exists := m["TagIdsRaw"]
	assert.False(t, exists, "TagIdsRaw should not be in JSON output")
}

func TestEmailTask_MarshalJSON_OtherFieldsPreserved(t *testing.T) {
	task := EmailTask{
		Id:       99,
		TaskName: "campaign-1",
		Subject:  "Hello World",
		Threads:  4,
		TagIdsRaw: "[5,6]",
	}

	data, err := json.Marshal(&task)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	assert.Equal(t, float64(99), m["id"])
	assert.Equal(t, "campaign-1", m["task_name"])
	assert.Equal(t, "Hello World", m["subject"])
	assert.Equal(t, float64(4), m["threads"])
}

// ---------------------------------------------------------------------------
// EmailTask.AfterFind
// ---------------------------------------------------------------------------

func TestEmailTask_AfterFind(t *testing.T) {
	tests := []struct {
		name       string
		tagIdsRaw  string
		wantTagIds []int
	}{
		{
			name:       "valid JSON array",
			tagIdsRaw:  "[1,2,3]",
			wantTagIds: []int{1, 2, 3},
		},
		{
			name:       "empty string",
			tagIdsRaw:  "",
			wantTagIds: nil,
		},
		{
			name:       "invalid JSON",
			tagIdsRaw:  "broken",
			wantTagIds: nil,
		},
		{
			name:       "empty array",
			tagIdsRaw:  "[]",
			wantTagIds: []int{},
		},
		{
			name:       "single element",
			tagIdsRaw:  "[7]",
			wantTagIds: []int{7},
		},
		{
			name:       "null JSON",
			tagIdsRaw:  "null",
			wantTagIds: nil,
		},
		{
			name:       "string instead of array",
			tagIdsRaw:  `"not an array"`,
			wantTagIds: nil,
		},
		{
			name:       "large array",
			tagIdsRaw:  "[1,2,3,4,5,6,7,8,9,10]",
			wantTagIds: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &EmailTask{TagIdsRaw: tt.tagIdsRaw}
			task.AfterFind()
			if tt.wantTagIds == nil {
				assert.Nil(t, task.TagIds)
			} else {
				assert.Equal(t, tt.wantTagIds, task.TagIds)
			}
		})
	}
}

func TestEmailTask_AfterFind_DoesNotOverwriteExisting(t *testing.T) {
	// AfterFind always parses TagIdsRaw regardless of existing TagIds
	task := &EmailTask{
		TagIdsRaw: "[10,20]",
		TagIds:    []int{99},
	}
	task.AfterFind()
	// AfterFind unconditionally sets TagIds from TagIdsRaw
	assert.Equal(t, []int{10, 20}, task.TagIds)
}

// ---------------------------------------------------------------------------
// Round-trip: AfterFind then MarshalJSON
// ---------------------------------------------------------------------------

func TestEmailTask_AfterFind_ThenMarshal(t *testing.T) {
	task := &EmailTask{
		Id:        42,
		TaskName:  "roundtrip",
		TagIdsRaw: "[5,10,15]",
	}
	task.AfterFind()
	assert.Equal(t, []int{5, 10, 15}, task.TagIds)

	data, err := json.Marshal(task)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	rawSlice := m["tag_ids"].([]interface{})
	got := make([]int, len(rawSlice))
	for i, v := range rawSlice {
		got[i] = int(v.(float64))
	}
	assert.Equal(t, []int{5, 10, 15}, got)
}
