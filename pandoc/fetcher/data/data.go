package data

import (
	"fmt"

	"github.com/gogap/config"
	"github.com/gogap/go-pandoc/pandoc/fetcher"
)

type DataFetcher struct {
}

type Params struct {
	Data []byte `json:"data"`
}

func (p *Params) Validation() (err error) {
	if len(p.Data) == 0 {
		err = fmt.Errorf("[fetcher-data]: params of data is empty")
		return
	}

	return
}

func init() {
	err := fetcher.RegisterFetcher("data", NewDataFetcher)

	if err != nil {
		panic(err)
	}
}

func NewDataFetcher(conf config.Configuration) (dataFetcher fetcher.Fetcher, err error) {
	dataFetcher = &DataFetcher{}
	return
}

func (p *DataFetcher) Fetch(fetchParams fetcher.FetchParams) (data []byte, err error) {

	params := Params{}

	err = fetchParams.Unmarshal(&params)
	if err != nil {
		return
	}

	err = params.Validation()
	if err != nil {
		return
	}

	data = params.Data

	return
}
