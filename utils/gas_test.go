package utils

import (
	"fmt"
	"testing"
)

func TestGetRequiredFeeFromError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		wantRequiredFee uint
		wantOk          bool
	}{
		{
			name:            "err is nil",
			err:             nil,
			wantRequiredFee: 0,
			wantOk:          false,
		},
		{
			name:            "valid pattern",
			err:             fmt.Errorf("error code: '13' msg: 'insufficient fees; got: 550amf required: 20419amf: insufficient fee'"),
			wantRequiredFee: 20419,
			wantOk:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRequiredFee, gotOk := GetRequiredFeeFromError(tt.err)
			if gotRequiredFee != tt.wantRequiredFee {
				t.Errorf("GetRequiredFeeFromError() gotRequiredFee = %v, want %v", gotRequiredFee, tt.wantRequiredFee)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetRequiredFeeFromError() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
