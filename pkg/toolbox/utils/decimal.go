package utils

import "github.com/shopspring/decimal"

func RoundDiv(a, b int64, round int32) float64 {
	res, _ := decimal.NewFromInt(a).Div(decimal.NewFromInt(b)).Round(round).Float64()
	return res
}
