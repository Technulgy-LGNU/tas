package web

import (
	"tas/internal/database"
	"testing"
)

func TestExtractShopDomain(t *testing.T) {
	tests := map[string]string{
		"https://www.example.com/product/1": "example.com",
		"shop.example.org/item":             "shop.example.org",
		"":                                  "",
	}

	for input, want := range tests {
		if got := extractShopDomain(input); got != want {
			t.Fatalf("extractShopDomain(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFuzzyInventoryMatchesRanksExactURLFirst(t *testing.T) {
	orderItem := database.OrderListItem{
		Name:       "Raspberry Pi 5",
		URL:        "https://shop.example.com/pi-5",
		ShopDomain: "shop.example.com",
	}
	inventory := []database.InventoryItem{
		{Base: database.Base{ID: 1}, Name: "Raspberry Pi 4", ProductURL: "https://other.example.com/pi-4"},
		{Base: database.Base{ID: 2}, Name: "Pi board", ProductURL: "https://shop.example.com/pi-5"},
	}

	matches := fuzzyInventoryMatches(orderItem, inventory)

	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Item.ID != 2 {
		t.Fatalf("top match ID = %d, want 2", matches[0].Item.ID)
	}
}
