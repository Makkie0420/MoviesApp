package app

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

var db *dynamodb.DynamoDB

// InitDynamoDB initializes the DynamoDB session and connects to the local DynamoDB instance
func InitDynamoDB() {
	// Connecting to local DynamoDB (make sure DynamoDB Local is running)
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("local"), // Use the local region (you can also use "us-west-2" but it's not critical in this case)
		Endpoint: aws.String("http://localhost:8000"), // Pointing to the local DynamoDB instance
	})
	if err != nil {
		log.Fatalf("Failed to connect to DynamoDB: %v", err)
	}
	db = dynamodb.New(sess)
}

// GetDB returns the DynamoDB client instance
func GetDB() *dynamodb.DynamoDB {
	return db
}

