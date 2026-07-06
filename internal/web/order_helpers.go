package web

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"tas/internal/database"
)

type inventoryMatch struct {
	Item  database.InventoryItem `json:"item"`
	Score int                    `json:"score"`
}

var nonWordPattern = regexp.MustCompile(`[^a-z0-9]+`)

func applyShop(domain *string, name *string, rawURL string, override string) {
	extracted := extractShopDomain(rawURL)
	*domain = extracted
	if strings.TrimSpace(override) != "" {
		*name = strings.TrimSpace(override)
		return
	}
	*name = extracted
}

func extractShopDomain(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("https://" + rawURL)
		if err != nil {
			return ""
		}
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

func normalizeSearchText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonWordPattern.ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(value), " ")
}

func fuzzyInventoryMatches(item database.OrderListItem, inventory []database.InventoryItem) []inventoryMatch {
	matches := make([]inventoryMatch, 0, len(inventory))
	needleName := normalizeSearchText(item.Name)
	needleURL := strings.TrimSpace(strings.ToLower(item.URL))

	for _, inv := range inventory {
		score := 0
		if needleURL != "" {
			productURL := strings.TrimSpace(strings.ToLower(inv.ProductURL))
			website := strings.TrimSpace(strings.ToLower(inv.Website))
			if productURL != "" && productURL == needleURL {
				score += 100
			}
			if website != "" && website == needleURL {
				score += 60
			}
			if extractShopDomain(inv.ProductURL) == item.ShopDomain || extractShopDomain(inv.Website) == item.ShopDomain {
				score += 20
			}
		}
		score += nameSimilarityScore(needleName, normalizeSearchText(inv.Name))
		if score > 0 {
			matches = append(matches, inventoryMatch{Item: inv, Score: score})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Item.Name < matches[j].Item.Name
		}
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > 8 {
		return matches[:8]
	}
	return matches
}

func nameSimilarityScore(a string, b string) int {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 80
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return 55
	}
	distance := levenshtein(a, b)
	maxLen := len([]rune(a))
	if other := len([]rune(b)); other > maxLen {
		maxLen = other
	}
	if maxLen == 0 {
		return 0
	}
	similarity := 100 - (distance * 100 / maxLen)
	if similarity < 35 {
		return 0
	}
	return similarity / 2
}

func levenshtein(a string, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	previous := make([]int, len(br)+1)
	current := make([]int, len(br)+1)
	for j := range previous {
		previous[j] = j
	}
	for i, ca := range ar {
		current[0] = i + 1
		for j, cb := range br {
			cost := 0
			if ca != cb {
				cost = 1
			}
			current[j+1] = minInt(
				current[j]+1,
				previous[j+1]+1,
				previous[j]+cost,
			)
		}
		previous, current = current, previous
	}
	return previous[len(br)]
}

func minInt(values ...int) int {
	best := values[0]
	for _, value := range values[1:] {
		if value < best {
			best = value
		}
	}
	return best
}
