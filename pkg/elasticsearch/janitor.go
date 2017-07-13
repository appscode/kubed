package es

import (
	"fmt"
	"time"

	"github.com/appscode/kubed/pkg/config"
	"github.com/appscode/log"
	elastic "gopkg.in/olivere/elastic.v3"
)

type Janitor struct {
	Config config.ClusterConfig
}

func (j *Janitor) CleanES() error {
	client, err := elastic.NewClient(
		// elastic.SetSniff(false),
		elastic.SetURL(j.Config.ElsticSearch.Endpoint),
	)
	if err != nil {
		return err
	}
	return j.DeleteIndices(client)
}

func (j *Janitor) DeleteIndices(client *elastic.Client) error {
	now := time.Now().UTC()
	oldDate := now.Add(time.Duration(-(j.Config.ElsticSearch.LogStorageLifetime)) * time.Second)

	// how many index should we check to delete? I set it to 7
	for i := 1; i <= 7; i++ {
		date := oldDate.AddDate(0, 0, -i)
		prefix := fmt.Sprintf("%s%s", j.Config.ElsticSearch.LogIndexPrefix, date.Format("2006.01.02"))

		if _, err := client.Search(prefix).Do(); err == nil {
			if _, err := client.DeleteIndex(prefix).Do(); err != nil {
				log.Errorln(err)
				return err
			}
			log.Debugf("Index [%s] deleted", prefix)
		}
	}
	log.Debugf("ElasticSearch cleanup process complete")
	return nil
}