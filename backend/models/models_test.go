package models

import "testing"

// A "table-driven test": the idiomatic way to test many cases in Go.
func TestProductInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   ProductInput
		wantErr bool
	}{
		{"valid", ProductInput{Name: "Mug", PriceCents: 500, Stock: 3}, false},
		{"empty name", ProductInput{Name: "   ", PriceCents: 500}, true},
		{"negative price", ProductInput{Name: "Mug", PriceCents: -1}, true},
		{"negative stock", ProductInput{Name: "Mug", Stock: -5}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
