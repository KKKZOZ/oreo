package txn

import (
	"testing"
)

// TestNewTransactionFactory is a unit test function that tests the behavior of the NewTransactionFactory function.
// It takes a testing.T parameter and runs multiple test cases to verify the correctness of NewTransactionFactory.
// Each test case includes a name, a configuration, an expected failure flag, and an expected error message.
// The function checks if NewTransactionFactory returns the expected error or no error based on the test case parameters.
func TestNewTransactionFactory(t *testing.T) {
	tests := []struct {
		name             string
		config           *TransactionConfig
		expectedToFail   bool
		expectedErrorMsg string
	}{
		{
			name:           "Nil config should initialize with default values",
			config:         nil,
			expectedToFail: false,
		},
		{
			name:             "Missing OracleURL when global time oracle source",
			config:           &TransactionConfig{TimeOracleSource: GLOBAL},
			expectedToFail:   true,
			expectedErrorMsg: "OracleURL is empty",
		},
		{
			name:             "Missing OracleURL when global locker source",
			config:           &TransactionConfig{LockerSource: GLOBAL},
			expectedToFail:   true,
			expectedErrorMsg: "OracleURL is empty",
		},
		{
			name:             "Inconsistent global time oracle and local locker mismatch",
			config:           &TransactionConfig{TimeOracleSource: GLOBAL, OracleURL: "http://localhost:8300", LockerSource: LOCAL},
			expectedToFail:   true,
			expectedErrorMsg: "LockerSource must be GLOBAL when using a global time oracle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransactionFactory(tt.config)
			if tt.expectedToFail {
				if err == nil {
					t.Error("expected error, got none")
				} else if err.Error() != tt.expectedErrorMsg {
					t.Errorf("expected error %q, got %q", tt.expectedErrorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}

// TestNewTransaction tests the creation of a new transaction using the TransactionFactory.
func TestNewTransaction(t *testing.T) {
	factory := TransactionFactory{
		TimeOracleSource: LOCAL,
		LockerSource:     LOCAL,
	}
	txn := factory.NewTransaction()

	if txn == nil {
		t.Errorf("expected created txn to be not nil")
	}
}
