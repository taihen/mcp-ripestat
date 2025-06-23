package types

import (
	"encoding/json"
	"testing"
)

func TestCustomTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     string
		wantErr  bool
	}{
		{
			name:     "RFC3339 format",
			jsonData: `"2024-05-20T18:05:00Z"`,
			want:     "2024-05-20 18:05:00 +0000 UTC",
			wantErr:  false,
		},
		{
			name:     "ISO format without timezone",
			jsonData: `"2024-05-20T18:05:00"`,
			want:     "2024-05-20 18:05:00 +0000 UTC",
			wantErr:  false,
		},
		{
			name:     "Space separated format",
			jsonData: `"2023-07-03 15:49:57"`,
			want:     "2023-07-03 15:49:57 +0000 UTC",
			wantErr:  false,
		},
		{
			name:     "Empty string",
			jsonData: `""`,
			want:     "0001-01-01 00:00:00 +0000 UTC",
			wantErr:  false,
		},
		{
			name:     "Null value",
			jsonData: `null`,
			want:     "0001-01-01 00:00:00 +0000 UTC",
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			jsonData: `"20-05-2024 18:05"`,
			want:     "0001-01-01 00:00:00 +0000 UTC",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct CustomTime
			err := json.Unmarshal([]byte(tt.jsonData), &ct)

			if (err != nil) != tt.wantErr {
				t.Errorf("CustomTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ct.Time.String() != tt.want {
				t.Errorf("CustomTime.UnmarshalJSON() = %v, want %v", ct.Time.String(), tt.want)
			}
		})
	}
}

func TestBaseResponse_Unmarshaling(t *testing.T) {
	jsonData := `{
		"messages": ["Test message"],
		"see_also": ["https://example.com"],
		"version": "1.0",
		"data_call_name": "test-call",
		"data_call_status": "supported",
		"cached": true,
		"query_id": "123456",
		"process_time": 42,
		"server_id": "app123",
		"build_version": "2024.06.23.1",
		"status": "ok",
		"status_code": 200,
		"time": "2024-06-23T11:30:00Z"
	}`

	var response BaseResponse
	err := json.Unmarshal([]byte(jsonData), &response)

	if err != nil {
		t.Fatalf("Failed to unmarshal BaseResponse: %v", err)
	}

	// Check a few fields to ensure proper unmarshaling

	if response.DataCallName != "test-call" {
		t.Errorf("Expected DataCallName to be 'test-call', got %q", response.DataCallName)
	}
	if response.StatusCode != 200 {
		t.Errorf("Expected StatusCode to be 200, got %d", response.StatusCode)
	}
	if response.ProcessTime != 42 {
		t.Errorf("Expected ProcessTime to be 42, got %d", response.ProcessTime)
	}
	if !response.Cached {
		t.Errorf("Expected Cached to be true, got %v", response.Cached)
	}
}

func TestBaseResponse_EmptyFields(t *testing.T) {
	// Test with minimal fields
	jsonData := `{
		"data_call_name": "test-call",
		"status": "ok",
		"status_code": 200
	}`

	var response BaseResponse
	err := json.Unmarshal([]byte(jsonData), &response)

	if err != nil {
		t.Fatalf("Failed to unmarshal minimal BaseResponse: %v", err)
	}

	// Check that fields were properly initialized
	if response.DataCallName != "test-call" {
		t.Errorf("Expected DataCallName to be 'test-call', got %q", response.DataCallName)
	}
	if response.StatusCode != 200 {
		t.Errorf("Expected StatusCode to be 200, got %d", response.StatusCode)
	}
	if len(response.Messages) != 0 {
		t.Errorf("Expected Messages to be empty, got %v", response.Messages)
	}
	if len(response.SeeAlso) != 0 {
		t.Errorf("Expected SeeAlso to be empty, got %v", response.SeeAlso)
	}
}
