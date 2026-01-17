package parser

import (
	"testing"
)

func TestParseRegex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantPhone string
		wantAddr  string
	}{
		{
			name:      "Standard Format v1",
			input:     "Name: John Doe\nPhone: 08012345678\nAddress: Lagos, Island\nCountry: Nigeria",
			wantName:  "John Doe",
			wantPhone: "08012345678",
			wantAddr:  "Lagos, Island",
		},
		{
			name:      "Standard Format v2",
			input:     "RECEIVER: Jane Smith\nTEL: +2347012345678\nADDR: Abuja Central\nCOUNTRY: NG",
			wantName:  "Jane Smith",
			wantPhone: "+2347012345678",
			wantAddr:  "Abuja Central",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRegex(tt.input)
			if got.ReceiverName != tt.wantName {
				t.Errorf("ParseRegex() Name = %v, want %v", got.ReceiverName, tt.wantName)
			}
			if got.ReceiverPhone != tt.wantPhone {
				t.Errorf("ParseRegex() Phone = %v, want %v", got.ReceiverPhone, tt.wantPhone)
			}
			if got.ReceiverAddress != tt.wantAddr {
				t.Errorf("ParseRegex() Address = %v, want %v", got.ReceiverAddress, tt.wantAddr)
			}
		})
	}
}
