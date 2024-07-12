package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/collection"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/amikos-tech/chroma-go/types"
)

// target_file = 'path/to/target/file'
// target_content = get_file_content(target_file)
// target_embedding = model.encode(target_content)

// # Query Chroma
// results = collection.query(target_embedding.tolist(), top_k=10)

// # Print results
// for result in results['matches']:
//     print(f"File: {result['id']}, Score: {result['score']}")

func setupChroma() {
	startServer()

	repoName, err := getRepoName()
	if err != nil {
		fmt.Println("Are you in a git repository?")
		return
	}

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

	newCollection, err := createOrGetCollection(client, repoName, openaiEf)
	if err != nil {
		fmt.Printf("error creating collection: %v", err)
		return
	}

	count, err := newCollection.Count(context.Background())
	if err != nil {
		fmt.Printf("error counting collection: %v", err)
		return
	}

	if count == 0 {
		err = createAndInsertRecords(newCollection, openaiEf)
		if err != nil {
			fmt.Printf("error creating and inserting records: %v", err)
			return
		}
	}

	err = queryCollection(newCollection)
	if err != nil {
		fmt.Printf("error querying collection: %v", err)
		return
	}
}

func initOpenAIEmbeddingFunction() (*openai.OpenAIEmbeddingFunction, error) {
	apiKey := openaiApiKey()
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
	client, err := chroma.NewClient("http://localhost:8000")
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

func getRepoName() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting remote URL: %v", err)
	}

	url := strings.TrimSpace(string(output))

	// Remove the scheme (http, https, git, etc.) and split the path
	var repoPath string
	if strings.HasPrefix(url, "git@") {
		// SSH format: git@github.com:username/repo.git
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid SSH URL format")
		}
		repoPath = parts[1]
	} else {
		// HTTPS format: https://github.com/username/repo.git
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid HTTPS URL format")
		}
		repoPath = strings.Join(parts[len(parts)-2:], "/")
	}

	// Remove the .git suffix if present
	repoPath = strings.TrimSuffix(repoPath, ".git")

	return repoPath, nil
}
