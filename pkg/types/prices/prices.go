package prices

const (
	SourceCoinGecko     = "coingecko"
	SourceBinance       = "binance"
	SourceKraken        = "kraken"
	SourceCryptoCompare = "cryptocompare"
	SourceDefiLlama     = "defillama"
	SourceGeckoTerminal = "geckoterminal"
)

type Asset struct {
	Name   string
	Symbol string
}

type Price struct {
	Asset       Asset
	Value       float64
	Source      string
	PoolAddress string
	Network     string
}

type PriceFetcher interface {
	Fetch(price *Price) error
	FetchMany(pairs ...*Price) error
	FetchAll() ([]Price, error)
}

var (
	SamplePrice = &Price{
		Asset: Asset{Name: "Bitcoin", Symbol: "BTC"},
	}
	SamplePrices = []*Price{
		{
			Asset: Asset{Name: "Bitcoin", Symbol: "BTC"},
		},
		{
			Asset: Asset{Name: "Ethereum", Symbol: "ETH"},
		},
	}
)
