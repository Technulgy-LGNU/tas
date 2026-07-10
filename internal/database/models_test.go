package database

import "testing"

func TestOrderRequestNormalizeTotalsCalculatesTotal(t *testing.T) {
	request := OrderRequest{Quantity: 3, UnitPriceCents: 1299}

	request.NormalizeTotals()

	if request.TotalPriceCents != 3897 {
		t.Fatalf("total = %d, want 3897", request.TotalPriceCents)
	}
}

func TestOrderRequestNormalizeTotalsNormalizesInvalidValues(t *testing.T) {
	request := OrderRequest{Quantity: 0, UnitPriceCents: -10}

	request.NormalizeTotals()

	if request.Quantity != 1 {
		t.Fatalf("quantity = %d, want 1", request.Quantity)
	}
	if request.UnitPriceCents != 0 {
		t.Fatalf("unit price = %d, want 0", request.UnitPriceCents)
	}
	if request.TotalPriceCents != 0 {
		t.Fatalf("total = %d, want 0", request.TotalPriceCents)
	}
}
