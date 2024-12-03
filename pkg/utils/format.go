package utils

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

func FormatAmount(amount float64, decimals int) string {
	var precision int
	switch {
	case amount >= 1000:
		precision = 2
	case amount >= 1:
		precision = 4
	case amount > 0:
		precision = 6
	default:
		precision = 2
	}

	if precision > decimals {
		precision = decimals
	}

	formatStr := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(formatStr, amount)
}

func ParseAmount(amount string, decimals int) float64 {
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0
	}
	return val / math.Pow10(decimals)
}

func ParseWeiToEther(wei *big.Int) float64 {
	f := new(big.Float)
	f.SetString(wei.String())
	ethValue := new(big.Float).Quo(f, big.NewFloat(1e18))
	result, _ := ethValue.Float64()
	return result
}
