package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/antavelos/blockchain/pkg/models/wallet"
)

func TestTransactionBody_getBalanceForAddress(t *testing.T) {
	testCases := []struct {
		name     string
		txb      TransactionBody
		address  string
		expected float64
	}{
		{
			name: "Recipient address",
			txb: TransactionBody{
				Recipient: "John",
				Sender:    "Jane",
				Amount:    100.0,
			},
			address:  "John",
			expected: 100.0,
		},
		{
			name: "Sender address",
			txb: TransactionBody{
				Recipient: "John",
				Sender:    "Jane",
				Amount:    100.0,
			},
			address:  "Jane",
			expected: -100.0,
		},
		{
			name: "Invalid address",
			txb: TransactionBody{
				Recipient: "John",
				Sender:    "Jane",
				Amount:    100.0,
			},
			address:  "Bob",
			expected: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.txb.getBalanceForAddress(tc.address)
			if got != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, got)
			}
		})
	}
}

func TestNewTransaction(t *testing.T) {
	// Creating sender and recipient wallets
	senderWallet, _ := wallet.NewWallet()
	recipientWallet, _ := wallet.NewWallet()

	// Creating new transaction with amount
	amount := 10.0
	tx, err := NewTransaction(*senderWallet, *recipientWallet, amount)
	if err != nil {
		t.Errorf("Failed to create new transaction: %v", err)
	}

	// Marshaling transaction body and verifying signature
	txbBytes, err := json.Marshal(tx.Body)
	if err != nil {
		t.Errorf("Failed to marshal transaction body: %v", err)
	}
	signatureBytes, _ := hex.DecodeString(tx.Signature)
	if !senderWallet.VerifySignature(txbBytes, signatureBytes) {
		t.Errorf("Invalid signature for transaction: %v", tx)
	}

	// Verifying transaction details
	if tx.Body.Sender != senderWallet.AddressString() {
		t.Errorf("Invalid sender for transaction: %v", tx)
	}
	if tx.Body.Recipient != recipientWallet.AddressString() {
		t.Errorf("Invalid recipient for transaction: %v", tx)
	}
	if tx.Body.Amount != amount {
		t.Errorf("Invalid amount for transaction: %v", tx)
	}
}

func TestIsCoinbase(t *testing.T) {
	tx := Transaction{
		Body: TransactionBody{
			Sender: "0",
		},
	}
	expectedResult := true
	result := tx.isCoinbase()
	if result != expectedResult {
		t.Errorf("isCoinbase() returned '%v', expected '%v'", result, expectedResult)
	}

	tx2 := Transaction{
		Body: TransactionBody{
			Sender: "random sender",
		},
	}
	expectedResult2 := false
	result2 := tx2.isCoinbase()
	if result2 != expectedResult2 {
		t.Errorf("isCoinbase() returned '%v', expected '%v'", result2, expectedResult2)
	}
}

// func TestTransaction_Validate(t *testing.T) {
// 	type fields struct {
// 		Body      *TransactionBody
// 		Signature string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid transaction",
// 			fields: fields{
// 				Body: &TransactionBody{
// 					Sender:    "sender",
// 					Recipient: "recipient",
// 					Amount:    100,
// 				},
// 				Signature: "a8d7e6b5c4",
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "invalid transaction with invalid signature",
// 			fields: fields{
// 				Body: &TransactionBody{
// 					Sender:    "sender",
// 					Recipient: "recipient",
// 					Amount:    100,
// 				},
// 				Signature: "invalid signature",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid transaction with invalid sender address",
// 			fields: fields{
// 				Body: &TransactionBody{
// 					Sender:    "invalid sender address",
// 					Recipient: "recipient",
// 					Amount:    100,
// 				},
// 				Signature: "a8d7e6b5c4",
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "invalid transaction with invalid body",
// 			fields: fields{
// 				Body:      nil,
// 				Signature: "a8d7e6b5c4",
// 			},
// 			wantErr: true,
// 		},
// 	}

//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				tx := Transaction{
//					Body:      *tt.fields.Body,
//					Signature: tt.fields.Signature,
//				}
//				if err := tx.Validate(); (err != nil) != tt.wantErr {
//					t.Errorf("Transaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
//				}
//			})
//		}
//	}
func TestHasTx(t *testing.T) {
	// Create a test block
	block := &Block{
		Txs: []Transaction{
			{Id: "tx1"},
			{Id: "tx2"},
		},
	}

	// Test a transaction that exists in the block
	tx1 := Transaction{Id: "tx1"}
	result1 := block.HasTx(tx1)
	if !result1 {
		t.Errorf("Expected HasTx to return true for existing transaction, but got false")
	}

	// Test a transaction that doesn't exist in the block
	tx3 := Transaction{Id: "tx3"}
	result2 := block.HasTx(tx3)
	if result2 {
		t.Errorf("Expected HasTx to return false for non-existing transaction, but got true")
	}
}
