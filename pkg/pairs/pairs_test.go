package pairs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPairPricing(t *testing.T) {
	pair := "TESTUSD"
	prices := testPrices()
	price, err := GetPriceForPair(pair, prices)
	assert.NoError(t, err)
	assert.Equal(t, 0.1, price)
}

func testPrices() map[string]float64 {
	return map[string]float64{
		"BTCUSDT": 100000.0,
		"ETHBTC":  0.1,
		"BNBETH":  0.1,
		"TRXBNB":  0.1,
		"XRPTRX":  0.1,
		"TESTXRP": 0.01,
	}
}
