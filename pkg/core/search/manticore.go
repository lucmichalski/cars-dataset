package search

import (
	"time"

	"github.com/manticoresoftware/go-sdk/manticore"
)

const (
	MANTICORE_HOST_DEFAULT string = `127.0.0.1`
	MANTICORE_PORT_DEFAULT uint16 = 9313
)

type ManticoreEngine struct {
	Client manticore.Client
}

func (me *ManticoreEngine) BatchIndex(documents []*Document) (int64, error) {
	start := time.Now().UnixNano() / int64(time.Millisecond)
	return time.Now().UnixNano()/int64(time.Millisecond) - start, nil
}

func (me *ManticoreEngine) Index(document *Document) (int64, error) {
	start := time.Now().UnixNano() / int64(time.Millisecond)
	return time.Now().UnixNano()/int64(time.Millisecond) - start, nil
}

func (me *ManticoreEngine) Search(query string) (interface{}, error) {
	return nil, nil
}

func (me *ManticoreEngine) Delete() error {
	return nil
}

func (me *ManticoreEngine) SetIndex(storeName string) error {
	return nil
}

func CreateManticoreClient(host string, port uint16) (manticore.Client, error) {
	cl := manticore.NewClient()
	cl.SetServer(host, port)
	cl.Open()
	return cl, nil
}