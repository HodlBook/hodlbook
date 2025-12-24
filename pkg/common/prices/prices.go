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
	SamplePair = &Price{
		Asset: Asset{Name: "Bitcoin", Symbol: "BTC"},
	}
	SamplePairs = []*Price{
		{
			Asset: Asset{Name: "Bitcoin", Symbol: "BTC"},
		},
		{
			Asset: Asset{Name: "Ethereum", Symbol: "ETH"},
		},
	}
)
