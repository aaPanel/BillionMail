package mail_service

import (
	"reflect"
	"testing"
)

func TestProcessForwardUsers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"newline split", "a@x.com\nb@x.com", []string{"a@x.com", "b@x.com"}},
		{"escaped newline split", `a@x.com\nb@x.com`, []string{"a@x.com", "b@x.com"}},
		{"trims whitespace", " a@x.com \n b@x.com ", []string{"a@x.com", "b@x.com"}},
		{"filters empty", "a@x.com\n\nb@x.com\n", []string{"a@x.com", "b@x.com"}},
		{"single address", "a@x.com", []string{"a@x.com"}},
		{"empty string", "", []string{}},
		{"only whitespace", "  \n  ", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessForwardUsers(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ProcessForwardUsers(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}
