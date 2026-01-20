package tests

import (
	"testing"
	"webtracker-bot/internal/parser"
)

func TestParseRegex(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantPhone  string
		wantAddr   string
		wantSender string
		wantEmail  string
		wantID     string
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
		{
			name:      "User Variation v1",
			input:     "Recivers's: name John Smith\nReceiver Phone: 08011122233\nAddress: 123 Street\nEmail: john@example.com\nID: A12345678",
			wantName:  "John Smith",
			wantPhone: "08011122233",
			wantAddr:  "123 Street",
			wantEmail: "john@example.com",
			wantID:    "A12345678",
		},
		{
			name:      "User Variation v2",
			input:     "reciver's name: Alice Cooper\nPhone: 09055566677\nReciver's Email: alice@gmail.com\nPassport: P9876543",
			wantName:  "Alice Cooper",
			wantPhone: "09055566677",
			wantEmail: "alice@gmail.com",
			wantID:    "P9876543",
		},
		{
			name:       "Sender Variation",
			input:      "Sender's Name: Bob Marley\nSender Country: Jamaica",
			wantSender: "Bob Marley",
		},
		{
			name:      "Misspelled Phone",
			input:     "Recivers's Phone: 08012344321",
			wantPhone: "08012344321",
		},
		{
			name:      "Recieve Variation",
			input:     "Recieve name: Charlie Brown\nRecieve's Mobile: 09011122233",
			wantName:  "Charlie Brown",
			wantPhone: "09011122233",
		},
		{
			name:      "Tel Variation",
			input:     "Reciver Tel: 08033344455",
			wantPhone: "08033344455",
		},
		{
			name:      "Mobil Variation",
			input:     "Resiver's Mobil: 07011122233",
			wantPhone: "07011122233",
		},
		{
			name:      "Receivers Variation",
			input:     "Receivers Name: John Doe\nReceivers Phone: 08012345678\nReceivers Address: 123 Main St",
			wantName:  "John Doe",
			wantPhone: "08012345678",
			wantAddr:  "123 Main St",
		},
		{
			name:       "Sender Plural Variation",
			input:      "Senders Name: Ziggy Stardust\nSenders Country: Mars",
			wantSender: "Ziggy Stardust",
		},
		{
			name:      "Capitalization and Whitespace",
			input:     "  RECEIVER'S NAME  :  Bruce Wayne  \n  PHONE  :  080-999-888  ",
			wantName:  "Bruce Wayne",
			wantPhone: "080-999-888",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.ParseRegex(tt.input)
			if tt.wantName != "" && got.ReceiverName != tt.wantName {
				t.Errorf("ParseRegex() Name = [%v], want [%v]", got.ReceiverName, tt.wantName)
			}
			if tt.wantPhone != "" && got.ReceiverPhone != tt.wantPhone {
				t.Errorf("ParseRegex() Phone = [%v], want [%v]", got.ReceiverPhone, tt.wantPhone)
			}
			if tt.wantAddr != "" && got.ReceiverAddress != tt.wantAddr {
				t.Errorf("ParseRegex() Address = [%v], want [%v]", got.ReceiverAddress, tt.wantAddr)
			}
			if tt.wantSender != "" && got.SenderName != tt.wantSender {
				t.Errorf("ParseRegex() SenderName = [%v], want [%v]", got.SenderName, tt.wantSender)
			}
			if tt.wantEmail != "" && got.ReceiverEmail != tt.wantEmail {
				t.Errorf("ParseRegex() Email = [%v], want [%v]", got.ReceiverEmail, tt.wantEmail)
			}
			if tt.wantID != "" && got.ReceiverID != tt.wantID {
				t.Errorf("ParseRegex() ID = [%v], want [%v]", got.ReceiverID, tt.wantID)
			}
		})
	}
}
