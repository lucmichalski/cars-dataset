package main

import (
	"github.com/gocolly/colly"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"strconv"
)

func main () {
	pagesToScan := 5

	c := colly.NewCollector(
		colly.AllowedDomains("www.autoreflex.com"),
		colly.MaxDepth(1),
	)

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		log.Panic(err)
	}

	// ./kafka-topics.sh --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic announces
	topic := "announces"

	defer producer.Close()

	go func () {
		for e := range producer.Events()  {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed : %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	} ()

	// A tr tag which has a star-id attribute corresponds to a car
	c.OnHTML("tr[star-id]", func(e *colly.HTMLElement){
		link := e.ChildAttr("a[href]", "href")
		// log.Printf("%s\n", link)

		err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value: []byte(link),
		}, nil)
		if err != nil {
			log.Printf("Error when producing %s : %s", link, err.Error())
		}
	})

	for i := 1; i <= pagesToScan; i ++ {
		err := c.Visit("http://www.autoreflex.com/137.0.-1.-1.-1.0.999999.1900.999999.-1.99.0." + strconv.Itoa(i) + "?fulltext=&geoban=M137R99")
		if err != nil {
			log.Panic(err)
		}
	}

	producer.Flush(15 * 1000)
}
