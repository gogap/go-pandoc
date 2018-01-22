package fetcher

import (
	"encoding/json"
	"fmt"

	"github.com/gogap/config"
)

type Fetcher interface {
	Fetch(FetchParams) ([]byte, error)
}

type FetchParams []byte

func (p *FetchParams) Unmarshal(v interface{}) (err error) {
	if p == nil {
		return
	}

	err = json.Unmarshal([]byte(*p), v)

	if err != nil {
		err = fmt.Errorf("parse param failure, error is %s", err.Error())
		return
	}

	return
}

type NewFetcherFunc func(config.Configuration) (Fetcher, error)

var (
	newFetcherFuncs = make(map[string]NewFetcherFunc)
)

func New(name string, conf config.Configuration) (f Fetcher, err error) {
	fn, exist := newFetcherFuncs[name]
	if !exist {
		err = fmt.Errorf("fetcher driver of %s not exist", name)
		return
	}

	return fn(conf)
}

func RegisterFetcher(name string, fn NewFetcherFunc) (err error) {

	if len(name) == 0 {
		err = fmt.Errorf("fetcher driver name is empty")
		return
	}

	if fn == nil {
		err = fmt.Errorf("the fetcher driver of %s's new func is nil", name)
		return
	}

	_, exist := newFetcherFuncs[name]

	if exist {
		err = fmt.Errorf("driver of %s already exist", name)
		return
	}

	newFetcherFuncs[name] = fn

	return
}
