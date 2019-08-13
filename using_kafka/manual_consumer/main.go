package main

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	topicName       = "test"
	bootstrapServer = "localhost:32771,localhost:32772"
)

// consumePartition consume specific partition datas in specific timespan
// offset that match ts must be set in partition before calling this function
// unit of ts and timeSpan is second
func consumePartition(partition kafka.TopicPartition, ts int64, timeSpan int) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServer,
		"group.id":           fmt.Sprintf("three_secs_consumerg_%d", ts),
		"enable.auto.commit": false,
	})

	if err != nil {
		fmt.Println("new consumer error")
		return
	}

	// assign partition manually will not subscribe rebalancing notification
	err = consumer.Assign([]kafka.TopicPartition{partition})
	if err != nil {
		fmt.Println("seek error", err)
		return
	}

L:
	for {
		msg, err := consumer.ReadMessage(time.Duration(timeSpan) * time.Second)
		if err == nil {
			mTs := msg.Timestamp.Unix()
			if mTs >= ts+int64(timeSpan) {
				fmt.Println(partition.Partition, "done")
				break
			}
			// Do what ever you want
			fmt.Printf("P[%d]: %d Message %s\n", partition.Partition, msg.Timestamp.Unix(), string(msg.Value))
		} else {
			switch e := err.(type) {
			case kafka.Error:
				if e.Code() == kafka.ErrTimedOut {
					fmt.Println(partition.Partition, "done")
					break L
				}

				fmt.Printf("P[%d] error: %s", partition.Partition, e)
				if e.IsFatal() {
					break L
				}
			default:
				fmt.Printf("P[%d] unkown error: %s", partition.Partition, err)
				break
			}
		}
	}

}

// getPartitionsWithBeginTime get partitions of a topic and set offsets
//   to offsets that match a timestamp
func getPartitionsWithBeginTime(tsInMillis int64) ([]kafka.TopicPartition, error) {
	timeout := 5000
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id":          "",
	})

	if err != nil {
		return nil, err
	}

	meta, err := consumer.GetMetadata(&topicName, true, timeout)
	if err != nil {
		return nil, err
	}

	topicMeta, ok := meta.Topics[topicName]
	if !ok {
		return nil, fmt.Errorf("topic not found")
	}

	partitionMetas := topicMeta.Partitions
	partitions := make([]kafka.TopicPartition, len(partitionMetas))
	for i := range partitions {
		partitions[i].Topic = &topicName
		partitions[i].Partition = partitionMetas[i].ID
		partitions[i].Offset = kafka.Offset(tsInMillis)
	}

	return consumer.OffsetsForTimes(partitions, timeout)
}

func main() {
	ts := int64(time.Now().Unix() - 9) // 9 seconds a go, just for example
	timeSpan := 3

	partitions, err := getPartitionsWithBeginTime(ts * 1000)
	if err != nil {
		fmt.Println("get partition error: ", err)
		return
	}

	for i := range partitions {
		fmt.Println("offset", i, "is", partitions[i].Offset)
	}

	var wg sync.WaitGroup
	for _, partition := range partitions {
		partition := partition
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumePartition(partition, ts, timeSpan)
		}()
	}

	wg.Wait()

}
