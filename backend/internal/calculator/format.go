package calculator

import (
	"math"
	"strconv"
	"strings"
)

func roundMoney(v float64) float64 {
	return math.Round(v)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// fmtAmount — целые рубли с разделителем тысяч пробелом: 275673 → "275 673".
func fmtAmount(v float64) string {
	s := strconv.FormatFloat(math.Abs(v), 'f', 0, 64)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
