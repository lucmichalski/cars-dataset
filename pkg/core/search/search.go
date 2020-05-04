package search

import (
	"errors"
	"strings"
	"strconv"
)

const (
	INDEX          		= "index"
	VENDOR_ELASTIC 		= "elastic"
	VENDOR_BLEVE   		= "bleve"
	VENDOR_MANTICORE    = "manticore"
)

type Document struct {
	Id   string
	Data []byte
	Query string
}

type SearchEngine interface {
	BatchIndex(documents []*Document) (int64, error)
	Index(document *Document) (int64, error)
	Search(query string) (interface{}, error)
	Delete() error
}

func GetSearchEngine(url *string, vendor *string, KVStore string) (SearchEngine, error) {
	var engine SearchEngine
	switch *vendor {
	case VENDOR_ELASTIC:
		// Create a client
		client, err := CreateElasticClient(url)
		if err != nil {
			return nil, err
		}
		engine = &ElasticEngine{client}

	case VENDOR_BLEVE:
		bleveEngine := &BleveEngine{}
		bleveEngine.SetKVStore(KVStore)
		engine = bleveEngine

	case VENDOR_MANTICORE:
		parts := strings.Split(*url, ":")
		if len(parts) < 2 {
			return nil, errors.New("Must provide hots:port url for manticore")			
		}
		host := parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		client, err := CreateManticoreClient(host, uint16(port))
		if err != nil {
			return nil, err
		}
		manticoreEngine := &ManticoreEngine{client}
		engine = manticoreEngine

	default:
		return nil, errors.New("Engine vendor must be specified.")
	}

	return engine, nil
}