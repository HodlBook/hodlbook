package coingeckoprices

import (
	"hodlbook/pkg/common/prices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceFetcher_Fetch(t *testing.T) {
	fetcher := NewPriceFetcher()
	err := fetcher.Fetch(prices.SamplePrice)
	assert.NoError(t, err)

	assert.NotZero(t, prices.SamplePrice.Value, "expected non-zero price value")
	t.Logf("%sUSD: %f", prices.SamplePrice.Asset.Symbol, prices.SamplePrice.Value)
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if err := fetcher.FetchMany(prices.SamplePrices...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, pair := range prices.SamplePrices {
		assert.NotZero(t, pair.Value, "expected non-zero price value for pair %v to USD", pair.Asset)
		t.Logf("%sUSD: %f", pair.Asset.Symbol, pair.Value)
	}
}
