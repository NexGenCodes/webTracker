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
		{
			name:      "Double Colon Separator",
			input:     "Receiver Name:: Clark Kent\nPhone:: 070-111-222",
			wantName:  "Clark Kent",
			wantPhone: "070-111-222",
		},
		{
			name:      "Hyphen Separator",
			input:     "Receiver-Name- Peter Parker\nPhone- 08012344321",
			wantName:  "Peter Parker",
			wantPhone: "08012344321",
		},
		{
			name:       "Mixed Case and Separators",
			input:      "reCeiVer nAMe: Tony Stark\nSENDER-NAME:: Steve Rogers\nITEM- box",
			wantName:   "Tony Stark",
			wantSender: "Steve Rogers",
		},
		{
			name:       "User Request - Spaces and Mixed Case",
			input:      "receiver Name: Peter Parker\nRECEIVER ADDRESS:: Queens\nreceiver phone - 555-0199\nSENDER Name: May Parker\nsender country: USA\nweight: 70",
			wantName:   "Peter Parker",
			wantAddr:   "Queens",
			wantPhone:  "555-0199",
			wantSender: "May Parker",
		},
		{
			name:      "User Request - Double Colon and Hyphen",
			input:     "receiver ID:: ID12345\nreceiver-Email- spidey@example.com",
			wantID:    "ID12345",
			wantEmail: "spidey@example.com",
		},
		{
			name:      "Same Line Mixed Fields",
			input:     "Receiver Name: Bruce Banner Phone: 555-0000 ID: HULK1",
			wantName:  "Bruce Banner",
			wantPhone: "555-0000",
			wantID:    "HULK1",
		},
		{
			name:     "Label as part of value",
			input:    "Receiver Name: John Address Lover\nAddress: Lagos",
			wantName: "John Address Lover",
			wantAddr: "Lagos",
		},
		{
			name:     "Missing labels",
			input:    "John Doe\n08012345678\nNigeria",
			wantName: "", // Currently parser requires a label or starts with 'name'
		},
		{
			name:     "Field swallows next label (name)",
			input:    "Receiver Address: 123 Main St name: Jane Doe",
			wantAddr: "123 Main St",
			wantName: "Jane Doe",
		},
		{
			name:      "Multi-line Address",
			input:     "Receiver Name: John Doe\nAddress: 123 Main St\nLagos Island\nNigeria\nPhone: 08012345678",
			wantName:  "John Doe",
			wantAddr:  "123 Main St\nLagos Island\nNigeria",
			wantPhone: "08012345678",
		},
		{
			name:      "Label-less Extraction (Fallback)",
			input:     "Jane Smith, Lagos, +2348011122233, john@example.com, weight 75.5",
			wantPhone: "+2348011122233",
			wantEmail: "john@example.com",
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
