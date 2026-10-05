package database

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
)

// OrderMoney stores thousandths of a cent (five decimal places in EUR).
// JSON and SQL retain the existing cents denomination, including fractional cents.
type OrderMoney int64

const Cent OrderMoney = 1000

var orderMoneyNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]{1,3})?$`)

func (m OrderMoney) MarshalJSON() ([]byte, error) {
	whole, fraction := m/Cent, m%Cent
	if fraction < 0 {
		fraction = -fraction
	}
	prefix := ""
	if m < 0 && whole == 0 {
		prefix = "-"
	}
	return []byte(fmt.Sprintf("%s%d.%03d", prefix, whole, fraction)), nil
}

func (m *OrderMoney) UnmarshalJSON(data []byte) error {
	// Bound parsing work and accept only numbers, never strings or null.
	if len(data) > 64 || !orderMoneyNumber.Match(data) {
		return fmt.Errorf("price must be a number of cents")
	}
	n, ok := new(big.Rat).SetString(string(data))
	if !ok {
		return fmt.Errorf("invalid price")
	}
	n.Mul(n, big.NewRat(int64(Cent), 1))
	if !n.IsInt() || !n.Num().IsInt64() {
		return fmt.Errorf("price must have at most five EUR decimal places and fit in int64")
	}
	*m = OrderMoney(n.Num().Int64())
	return nil
}

func (m OrderMoney) Value() (driver.Value, error) {
	data, err := m.MarshalJSON()
	return string(data), err
}

func (m *OrderMoney) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return m.UnmarshalJSON([]byte(v))
	case []byte:
		return m.UnmarshalJSON(v)
	case int64:
		return m.UnmarshalJSON([]byte(strconv.FormatInt(v, 10)))
	default:
		return fmt.Errorf("unsupported price database value %T", value)
	}
}
