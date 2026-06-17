package jsonutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "number", raw: `20`, want: 20},
		{name: "string", raw: `"20"`, want: 20},
		{name: "null", raw: `null`, want: 0},
		{name: "empty string", raw: `""`, want: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got FlexInt
			require.NoError(t, json.Unmarshal([]byte(tt.raw), &got))
			assert.Equal(t, tt.want, got.Int())
		})
	}
}

func TestFlexString_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "string", raw: `"2"`, want: "2"},
		{name: "number", raw: `0`, want: "0"},
		{name: "null", raw: `null`, want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got FlexString
			require.NoError(t, json.Unmarshal([]byte(tt.raw), &got))
			assert.Equal(t, tt.want, got.String())
		})
	}
}
