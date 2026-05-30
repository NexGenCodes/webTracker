package main

import (
	"fmt"
	"webtracker-bot/internal/parser"
)

func main() {
	// Test Case 1: Structured manifest with postal code and county
	text1 := `RECEIVER'S COUNTRY: Spain

RECEIVER'S NAME: Carlos Alberto santa Hernández

RECEIVER'S ADDRESS: Carrera 16 numero 16-50 apartamento 293 

RECEIVER'S POSTAL CODE: 08016

RECEIVER'S NUMBER: +34 614 42 74 34

SENDER'S NAME: Juliana 

SENDER'S COUNTY: IRAN`

	fmt.Println("--- Test Case 1 (Structured with Postal Code/County) ---")
	m1 := parser.ParseRegex(text1)
	fmt.Printf("ReceiverName: %q\n", m1.ReceiverName)
	fmt.Printf("ReceiverAddress: %q\n", m1.ReceiverAddress)
	fmt.Printf("ReceiverCountry: %q\n", m1.ReceiverCountry)
	fmt.Printf("ReceiverPhone: %q\n", m1.ReceiverPhone)
	fmt.Printf("SenderName: %q\n", m1.SenderName)
	fmt.Printf("SenderCountry: %q\n", m1.SenderCountry)

	// Test Case 2: Simple manifest with only "RECEIVER:" and "SENDER:"
	text2 := `RECEIVER: Carlos Alberto
SENDER: Juliana
COUNTRY: Spain`

	fmt.Println("\n--- Test Case 2 (Simple receiver/sender without word 'name') ---")
	m2 := parser.ParseRegex(text2)
	fmt.Printf("ReceiverName: %q\n", m2.ReceiverName)
	fmt.Printf("SenderName: %q\n", m2.SenderName)
}


