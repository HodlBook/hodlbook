package prices

type Asset struct {
	Name   string
	Symbol string
}

type Price struct {
	FromAsset Asset
	Value     uint64
}

type PriceFetcher interface {
	Fetch(price *Price) error
	FetchMany(pairs ...*Price) error
}

var (
	SamplePair = &Price{
		FromAsset: Asset{Name: "Bitcoin", Symbol: "BTC"},
	}
	SamplePairs = []*Price{
		{
			FromAsset: Asset{Name: "Bitcoin", Symbol: "BTC"},
		},
		{
			FromAsset: Asset{Name: "Ethereum", Symbol: "ETH"},
		},
	}
)
