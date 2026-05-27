package valueobject

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
)

type MoneyAmountCurrency string

const MoneyAmountCurrencyBYN MoneyAmountCurrency = "BYN"

type MoneyAmount struct {
	amount    int64
	decimals  uint8
	currency  MoneyAmountCurrency
	precision *uint8
}

func NewMoneyAmountFromAtomic(amountAtomic string, decimals uint8, currency MoneyAmountCurrency, precision *uint8) (MoneyAmount, error) {
	if currency == "" {
		return MoneyAmount{}, errors.New("empty currency")
	}

	i := new(big.Int)
	_, ok := i.SetString(amountAtomic, 10)
	if !ok {
		return MoneyAmount{}, errors.New("invalid atomic amount")
	}

	if i.Sign() < 0 {
		return MoneyAmount{}, errors.New("amount cannot be negative")
	}

	if !i.IsInt64() {
		return MoneyAmount{}, errors.New("amount overflow")
	}

	return MoneyAmount{
		amount:    i.Int64(),
		decimals:  decimals,
		currency:  currency,
		precision: precision,
	}, nil
}

func NewMoneyAmountFromInt64(amountAtomic int64, decimals uint8, currency MoneyAmountCurrency, precision *uint8) (MoneyAmount, error) {
	if currency == "" {
		return MoneyAmount{}, errors.New("empty currency")
	}

	if amountAtomic < 0 {
		return MoneyAmount{}, errors.New("amount cannot be negative")
	}

	return MoneyAmount{
		amount:    amountAtomic,
		decimals:  decimals,
		currency:  currency,
		precision: precision,
	}, nil
}

func NewZeroMoneyAmount(decimals uint8, currency MoneyAmountCurrency, precision *uint8) (MoneyAmount, error) {
	if currency == "" {
		return MoneyAmount{}, errors.New("empty currency")
	}

	return MoneyAmount{
		amount:    0,
		decimals:  decimals,
		currency:  currency,
		precision: precision,
	}, nil
}

func NewMoneyAmountFromDecimal(amount string, decimals uint8, currency MoneyAmountCurrency, precision *uint8) (MoneyAmount, error) {
	if currency == "" {
		return MoneyAmount{}, errors.New("empty currency")
	}

	if amount == "" {
		return MoneyAmount{}, errors.New("empty amount")
	}

	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return MoneyAmount{}, errors.New("invalid decimal format")
	}

	intPart := parts[0]
	fracPart := ""

	if len(parts) == 2 {
		fracPart = parts[1]
	}

	if len(fracPart) > int(decimals) {
		return MoneyAmount{}, errors.New("too many fractional digits")
	}

	fracPart = fracPart + strings.Repeat("0", int(decimals)-len(fracPart))

	full := intPart + fracPart
	full = strings.TrimLeft(full, "0")
	if full == "" {
		full = "0"
	}

	i := new(big.Int)
	_, ok := i.SetString(full, 10)
	if !ok {
		return MoneyAmount{}, errors.New("invalid decimal amount")
	}

	if !i.IsInt64() {
		return MoneyAmount{}, errors.New("amount overflow")
	}

	return MoneyAmount{
		amount:    i.Int64(),
		decimals:  decimals,
		currency:  currency,
		precision: precision,
	}, nil
}

func (a MoneyAmount) Equal(other MoneyAmount) bool {
	if a.decimals != other.decimals || a.currency != other.currency {
		return false
	}
	return a.amount == other.amount
}

func (a MoneyAmount) Compare(other MoneyAmount) (int, error) {
	if a.decimals != other.decimals {
		return 0, errors.New("decimals mismatch")
	}
	if a.currency != other.currency {
		return 0, errors.New("currency mismatch")
	}
	switch {
	case a.amount < other.amount:
		return -1, nil
	case a.amount > other.amount:
		return 1, nil
	default:
		return 0, nil
	}
}

func (a MoneyAmount) GreaterThan(other MoneyAmount) (bool, error) {
	cmp, err := a.Compare(other)
	return cmp > 0, err
}

func (a MoneyAmount) GreaterThanOrEqual(other MoneyAmount) (bool, error) {
	cmp, err := a.Compare(other)
	return cmp >= 0, err
}

func (a MoneyAmount) LessThan(other MoneyAmount) (bool, error) {
	cmp, err := a.Compare(other)
	return cmp < 0, err
}

func (a MoneyAmount) LessThanOrEqual(other MoneyAmount) (bool, error) {
	cmp, err := a.Compare(other)
	return cmp <= 0, err
}

func (a MoneyAmount) Add(other MoneyAmount) (MoneyAmount, error) {
	if a.decimals != other.decimals {
		return MoneyAmount{}, errors.New("decimals mismatch")
	}
	if a.currency != other.currency {
		return MoneyAmount{}, errors.New("currency mismatch")
	}

	if other.amount > 0 && a.amount > math.MaxInt64-other.amount {
		return MoneyAmount{}, errors.New("integer overflow")
	}

	return MoneyAmount{
		amount:    a.amount + other.amount,
		decimals:  a.decimals,
		currency:  a.currency,
		precision: a.precision,
	}, nil
}

func (a MoneyAmount) Sub(other MoneyAmount) (MoneyAmount, error) {
	if a.decimals != other.decimals {
		return MoneyAmount{}, errors.New("decimals mismatch")
	}
	if a.currency != other.currency {
		return MoneyAmount{}, errors.New("currency mismatch")
	}

	result := a.amount - other.amount
	if result < 0 {
		return MoneyAmount{}, errors.New("negative result")
	}

	return MoneyAmount{
		amount:    result,
		decimals:  a.decimals,
		currency:  a.currency,
		precision: a.precision,
	}, nil
}

func (a MoneyAmount) MulFloat64(quantity float64) (MoneyAmount, error) {
	if quantity < 0 {
		return MoneyAmount{}, errors.New("negative quantity")
	}
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return MoneyAmount{}, errors.New("invalid quantity")
	}
	if a.amount == 0 || quantity == 0 {
		return MoneyAmount{
			amount:    0,
			decimals:  a.decimals,
			currency:  a.currency,
			precision: a.precision,
		}, nil
	}

	amtRat := new(big.Rat).SetInt64(a.amount)
	qRat := new(big.Rat).SetFloat64(quantity)

	res := new(big.Rat).Mul(amtRat, qRat)
	rounded, err := roundRatToInt64HalfUp(res)
	if err != nil {
		return MoneyAmount{}, err
	}

	return MoneyAmount{
		amount:    rounded,
		decimals:  a.decimals,
		currency:  a.currency,
		precision: a.precision,
	}, nil
}

func roundRatToInt64HalfUp(r *big.Rat) (int64, error) {
	if r.Sign() < 0 {
		return 0, errors.New("negative result")
	}

	num := r.Num()
	den := r.Denom()

	var quot, rem big.Int
	quot.DivMod(num, den, &rem)

	twiceRem := new(big.Int).Lsh(&rem, 1)
	if twiceRem.Cmp(den) >= 0 {
		quot.Add(&quot, big.NewInt(1))
	}

	if !quot.IsInt64() {
		return 0, errors.New("integer overflow")
	}

	return quot.Int64(), nil
}

func (a MoneyAmount) AtomicString() string {
	return strconv.FormatInt(a.amount, 10)
}

func (a MoneyAmount) DecimalString() string {
	s := a.AtomicString()

	if a.decimals == 0 {
		return s
	}

	if len(s) <= int(a.decimals) {
		s = strings.Repeat("0", int(a.decimals)-len(s)+1) + s
	}

	intPart := s[:len(s)-int(a.decimals)]
	fracPart := s[len(s)-int(a.decimals):]

	fracPart = strings.TrimRight(fracPart, "0")

	if fracPart == "" {
		return intPart
	}

	return intPart + "." + fracPart
}

func (a MoneyAmount) String() string {
	dec := a.DecimalString()

	if a.precision == nil {
		return dec
	}

	p := int(*a.precision)

	parts := strings.Split(dec, ".")
	intPart := parts[0]

	if len(parts) == 1 {
		if p == 0 {
			return intPart
		}
		return intPart + "." + strings.Repeat("0", p)
	}

	frac := parts[1]

	if len(frac) < p {
		return intPart + "." + frac + strings.Repeat("0", p-len(frac))
	}

	if len(frac) == p {
		return intPart + "." + frac
	}

	roundDigit := frac[p]
	frac = frac[:p]

	if roundDigit < '5' {
		return intPart + "." + frac
	}

	full := intPart + frac

	i := new(big.Int)
	i.SetString(full, 10)
	i.Add(i, big.NewInt(1))

	result := i.String()

	if len(result) <= p {
		result = strings.Repeat("0", p-len(result)+1) + result
	}

	intPart = result[:len(result)-p]
	frac = result[len(result)-p:]

	return intPart + "." + frac
}

func (a MoneyAmount) Decimals() uint8 {
	return a.decimals
}

func (a MoneyAmount) Currency() string {
	return string(a.currency)
}

func (a MoneyAmount) Precision() (uint8, bool) {
	if a.precision == nil {
		return 0, false
	}
	return *a.precision, true
}

func (a MoneyAmount) Int64() int64 {
	return a.amount
}

func (a MoneyAmount) BigInt() *big.Int {
	return big.NewInt(a.amount)
}

func (a MoneyAmount) WithPrecision(precision uint8) MoneyAmount {
	a.precision = &precision
	return a
}

func (a MoneyAmount) ClearPrecision() MoneyAmount {
	a.precision = nil
	return a
}
