package coingeckoprices

import (
	"hodlbook/pkg/common/prices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceFetcher_Fetch(t *testing.T) {
	fetcher := NewPriceFetcher()
	err := fetcher.Fetch(prices.SamplePair)
	assert.NoError(t, err)

	assert.NotZero(t, prices.SamplePair.Value, "expected non-zero price value")
	t.Logf("%sUSD: %d", prices.SamplePair.FromAsset.Symbol, prices.SamplePair.Value)
}

func TestPriceFetcher_FetchMany(t *testing.T) {
	fetcher := NewPriceFetcher()

	if err := fetcher.FetchMany(prices.SamplePairs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, pair := range prices.SamplePairs {
		assert.NotZero(t, pair.Value, "expected non-zero price value for pair %v to USD", pair.FromAsset)
		t.Logf("%sUSD: %d", pair.FromAsset.Symbol, pair.Value)
	}
}
