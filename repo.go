package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/collection"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/amikos-tech/chroma-go/types"
)

func runTest() {
	openaiEf, err := initOpenAIEmbeddingFunction()
	if err != nil {
		fmt.Printf("error initializing OpenAI embedding function: %v", err)
		return
	}

	client := createChromaClient()
	if client == nil {
		fmt.Println("error creating client")
		return
	}

	newCollection, err := createOrGetCollection(client, "test_collection", openaiEf)
	if err != nil {
		fmt.Printf("error creating collection: %v", err)
		return
	}

	err = createAndInsertRecords(newCollection, openaiEf)
	if err != nil {
		fmt.Printf("error creating and inserting records: %v", err)
		return
	}

	err = queryCollection(newCollection)
	if err != nil {
		fmt.Printf("error querying collection: %v", err)
		return
	}
}

func initOpenAIEmbeddingFunction() (*openai.OpenAIEmbeddingFunction, error) {
	apiKey := apiKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no key found")
	}
	openaiEf, err := openai.NewOpenAIEmbeddingFunction(apiKey)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAI embedding function: %v", err)
	}
	return openaiEf, nil
}

func createChromaClient() *chroma.Client {
	client, err := chroma.NewClient("http://localhost:53829")
	if err != nil {
		fmt.Printf("error creating client: %v", err)
		return nil
	}
	return client
}

func createOrGetCollection(client *chroma.Client, collectionName string, openaiEf *openai.OpenAIEmbeddingFunction) (*chroma.Collection, error) {
	c, err := client.GetCollection(context.TODO(), collectionName, openaiEf)
	if err == nil {
		return c, nil
	}

	newCollection, err := client.NewCollection(
		context.TODO(),
		collection.WithName(collectionName),
		collection.WithMetadata("key1", "value1"),
		collection.WithEmbeddingFunction(openaiEf),
		collection.WithHNSWDistanceFunction(types.L2),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating collection: %v", err)
	}
	return newCollection, nil
}

func createAndInsertRecords(newCollection *chroma.Collection, openaiEf *openai.OpenAIEmbeddingFunction) error {
	rs, err := types.NewRecordSet(
		types.WithEmbeddingFunction(openaiEf),
		types.WithIDGenerator(types.NewULIDGenerator()),
	)
	if err != nil {
		return fmt.Errorf("error creating record set: %v", err)
	}

	rs.WithRecord(types.WithDocument("My name is John. And I have two dogs."), types.WithMetadata("key1", "value1"))
	rs.WithRecord(types.WithDocument("My name is Jane. I am a data scientist."), types.WithMetadata("key2", "value2"))

	_, err = rs.BuildAndValidate(context.TODO())
	if err != nil {
		return fmt.Errorf("error validating record set: %v", err)
	}

	_, err = newCollection.AddRecords(context.Background(), rs)
	if err != nil {
		return fmt.Errorf("error adding documents: %v", err)
	}

	return nil
}

func queryCollection(newCollection *chroma.Collection) error {
	countDocs, err := newCollection.Count(context.TODO())
	if err != nil {
		return fmt.Errorf("error counting documents: %v", err)
	}
	fmt.Printf("countDocs: %v\n", countDocs) // this should result in 2

	qr, err := newCollection.Query(context.TODO(), []string{"I love dogs"}, 5, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("error querying documents: %v", err)
	}
	fmt.Printf("qr: %v\n", qr.Documents[0][0]) // this should result in the document about dogs

	return nil
}

func isChromaRunning(port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func startChroma() error {
	cmd := exec.Command("chroma", "--config", "/path/to/config.yaml")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting Chroma: %v", err)
	}
	time.Sleep(2 * time.Second) // Give Chroma some time to start
	return nil
}
