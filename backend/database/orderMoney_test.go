package database

import (
	"encoding/json"
	"testing"
)

func TestOrderMoneyPrecision(t *testing.T) {
	for _, tc := range []struct {
		json string
		want OrderMoney
	}{
		{"199", 199 * Cent}, // Existing JSON prices retain their value.
		{"0", 0},
		{"0.001", 1}, // EUR 0.00001
		{"0.123", 123},
		{"100000000.000", 100000000 * Cent},
		{"-0.001", -1}, // Decoded exactly, then rejected by price validation.
	} {
		t.Run(tc.json, func(t *testing.T) {
			var got OrderMoney
			if err := json.Unmarshal([]byte(tc.json), &got); err != nil || got != tc.want {
				t.Fatalf("decode: %d, %v", got, err)
			}
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var roundTrip OrderMoney
			if err := json.Unmarshal(encoded, &roundTrip); err != nil || roundTrip != got {
				t.Fatalf("JSON round trip: %s, %v", encoded, err)
			}
			value, err := got.Value()
			if err != nil {
				t.Fatal(err)
			}
			if err := roundTrip.Scan(value); err != nil || roundTrip != got {
				t.Fatalf("SQL round trip: %v, %v", value, err)
			}
		})
	}
	for _, invalid := range []string{`0.0001`, `"0.001"`, `null`, `1e999999999`, `999999999999999999999999`, `1/2`} {
		var price OrderMoney
		if err := json.Unmarshal([]byte(invalid), &price); err == nil {
			t.Fatalf("accepted invalid price %s", invalid)
		}
	}
	var legacy OrderMoney
	if err := legacy.Scan(int64(199)); err != nil || legacy != 199*Cent {
		t.Fatalf("legacy SQL cents changed: %d, %v", legacy, err)
	}
}
