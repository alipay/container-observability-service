package xsearch

import "github.com/olivere/elastic/v7"

func GetXSearchClient() *elastic.Client {
	return esClient
}
