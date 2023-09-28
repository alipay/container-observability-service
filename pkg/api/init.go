package api

import "github.com/olivere/elastic"

var (
	esClient *elastic.Client
)

func InitApi(endPoint string, username string, password string) {
	var err error
	esClient, err = elastic.NewClient(
		elastic.SetURL(endPoint), elastic.SetBasicAuth(username, password), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
}
