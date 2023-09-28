package xsearch

import "github.com/olivere/elastic"

func GetXSearchClient() *elastic.Client {
	return esClient
}
