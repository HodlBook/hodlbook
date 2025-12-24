package pairs

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrPairNotFound = errors.New("price for pair not found")
	// ordered by pair count desc
	//
	// as fetched from Binance API
	anchor = []string{
		"USDT",
		"USD",
		"BTC",
		"TRY",
		"USDC",
		"BNB",
		"ETH",
		"EUR",
		"IDR",
		"BRL",
		"JPY",
		"RUB",
		"GBP",
		"AUD",
		"USD1",
		"UAH",
		"PLN",
		"ARS",
		"DAI",
		"MXN",
		"RON",
		"DRT",
		"ZAR",
		"EURI",
		"USDP",
		"CZK",
		"NGN",
		"VAI",
		"BVND",
		"XRP",
		"SOL",
		"DOT",
		"COP",
		"DOGE",
		"TRX",
	}
)

type Pair struct {
	Price  string `json:"price"`
	Symbol string `json:"symbol"`
}

func GetPriceForPair(pair string, prices map[string]float64) (float64, error) {
	// Base case: if the pair is USDT, return 1:1
	if pair == "USDUSDT" {
		return 1.0, nil
	}

	// Check if the pair exists in the prices map
	if price, exists := prices[pair]; exists {
		return price, nil
	}

	anchorPairs := map[string]map[string]struct{}{}
	base := ""
	for _, anchor := range anchor {
		if strings.HasSuffix(pair, anchor) {
			base = strings.TrimSuffix(pair, anchor)
		}
		for pair := range prices {
			if strings.HasSuffix(pair, anchor) {
				if _, exists := anchorPairs[anchor]; !exists {
					anchorPairs[anchor] = map[string]struct{}{}
				}
				anchorPairs[anchor][strings.TrimSuffix(pair, anchor)] = struct{}{}
			}
		}
	}
	if base == "" {
		return 0, errors.Wrap(ErrPairNotFound, pair)
	}
	for anchor, assets := range anchorPairs {
		if _, exists := assets[base]; exists {
			priceBaseAnchor, err := GetPriceForPair(base+anchor, prices)
			if err != nil {
				return 0, err
			}
			if priceBaseAnchor == 0 {
				continue
			}
			priceAnchorUSDT, err := GetPriceForPair(anchor+"USDT", prices)
			if err != nil {
				return 0, err
			}
			return priceBaseAnchor * priceAnchorUSDT, nil
		}
	}
	return 0, errors.Wrap(ErrPairNotFound, pair)
}

func PairMap(pairs []Pair) map[string]float64 {
	priceMap := make(map[string]float64)
	for _, pair := range pairs {
		priceMap[pair.Symbol] = parsePrice(pair.Price)
	}
	return priceMap
}

func parsePrice(priceStr string) float64 {
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0.0
	}
	return price
}
