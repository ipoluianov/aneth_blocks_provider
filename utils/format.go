package utils

import (
	"fmt"
	"math"
	"math/big"
)

func FormatIntString(str string) string {
	res := make([]byte, 0)
	for i := 0; i < len(str); i++ {
		k := 3 - (len(str) % 3)
		if ((i + k) % 3) == 0 {
			res = append(res, ' ')
		}
		res = append(res, str[i])
	}
	return string(res)
}

func FormarValueToGWEI(value *big.Int) string {
	castFloat := new(big.Float)
	castFloat.SetInt(value)

	ethValue := new(big.Float).Quo(castFloat, big.NewFloat(math.Pow10(9)))
	ethValueInt64, _ := ethValue.Int64()

	//bi.Cast = ethValue.String()
	valueStr := fmt.Sprint(ethValueInt64)

	valueStr = FormatIntString(valueStr) + " GWEI"
	return valueStr
}
