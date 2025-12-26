package prices

type Asset struct {
	Name   string
	Symbol string
}

type Price struct {
	Asset Asset
	Value float64
}

type PriceFetcher interface {
	Fetch(price *Price) error
	FetchMany(pairs ...*Price) error
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
