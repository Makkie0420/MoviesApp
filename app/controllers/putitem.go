package controllers

import (
	"MoviesApp/app/models"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func InsertMovie(movie models.Movie) error {
	// Initialize the DynamoDB client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), //  region
	})
	if err != nil {
		return fmt.Errorf("unable to create DynamoDB session: %v", err)
	}

	svc := dynamodb.New(sess)

	// Prepare the movie data to be inserted
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Movies"),
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(movie.ID),
			},
			"Title": {
				S: aws.String(movie.Title),
			},
			"Plot": {
				S: aws.String(movie.Plot),
			},
			"Year": {
				S: aws.String(movie.Year),
			},
			"Rating": {
				N: aws.String(fmt.Sprintf("%f", movie.Rating)),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to insert movie: %v", err)
	}

	return nil
}
