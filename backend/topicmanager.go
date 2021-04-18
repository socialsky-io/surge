package surge

import (
	"encoding/json"
	"log"

	"github.com/rule110-io/surge/backend/models"
	"github.com/rule110-io/surge/backend/mutexes"
)

var topicsMap map[string]models.Topic

const topicsMapBucketKey = "topicBucket"

func InitializeTopicsManager() {
	topicsMap = make(map[string]models.Topic)

	//Load from db
	mapString, err := DbReadSetting(topicsMapBucketKey)
	if err == nil {
		mapBytes := []byte(mapString)
		err := json.Unmarshal(mapBytes, &topicsMap)
		if err != nil {
			log.Println("Failed to unmarshal setting for topics", err)
		}
	} else {
		log.Println("Failed to read setting for topics", err)
	}
}

func subscribeToSurgeTopic(topicName string) {
	mutexes.TopicsMapLock.Lock()
	defer mutexes.TopicsMapLock.Unlock()

	if _, ok := topicsMap[topicName]; ok {
		//Already subscribed to this topic
		return
	}

	topicEncoded := TopicEncode(topicName)

	topicModel := models.Topic{
		Name:        topicName,
		NameEncoded: topicEncoded,
	}

	topicsMap[topicName] = topicModel

	//Save to our bucket
	mapBytes, err := json.Marshal(topicsMap)
	if err == nil {
		mapString := string(mapBytes)
		DbWriteSetting(topicsMapBucketKey, mapString)
	}

	subscribeToPubSub(topicEncoded)
}

func unsubscribeFromSurgeTopic(topicName string) {
	mutexes.TopicsMapLock.Lock()
	defer mutexes.TopicsMapLock.Unlock()

	if topic, ok := topicsMap[topicName]; ok {
		unsubscribeToPubSub(topic.NameEncoded)
	}

	//Delete from map
	delete(topicsMap, topicName)

	//Save to our bucket
	mapBytes, err := json.Marshal(topicsMap)
	if err == nil {
		mapString := string(mapBytes)
		DbWriteSetting(topicsMapBucketKey, mapString)
	}
}

func resubscribeToTopics() {
	mutexes.TopicsMapLock.Lock()
	defer mutexes.TopicsMapLock.Unlock()
	for _, topic := range topicsMap {
		subscribeToPubSub(topic.NameEncoded)
	}
}