package controllers

import (
	"MoviesApp/app"
	"MoviesApp/app/models"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oklog/ulid/v2"
	"github.com/revel/revel"
)

type MovieController struct {
	*revel.Controller
}

// generate ULID for unique ID
func generateULID() string {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

// GET /movies

func (c MovieController) ListMovies() revel.Result {
	svc := app.GetDB()

	input := &dynamodb.ScanInput{
		TableName: aws.String("Movies"),
	}

	result, err := svc.Scan(input)
	if err != nil {
		c.Response.Status = 500
		return c.RenderText("Failed to scan movies")
	}

	var movies []models.Movie
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &movies)
	if err != nil {
		c.Response.Status = 500
		return c.RenderText("Failed to unmarshal movies")
	}

	return c.RenderJSON(movies)
}

// GET /movies/{movieId}
func (c MovieController) GetMovie(movieId string) revel.Result {
	svc := app.GetDB()

	params := &dynamodb.GetItemInput{
		TableName: aws.String("Movies"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(movieId)},
		},
	}

	result, err := svc.GetItem(params)
	if err != nil {
		revel.AppLog.Errorf("Failed to fetch movie: %v", err)
		c.Response.Status = 500
		return c.RenderText("Failed to fetch movie")
	}

	if result.Item == nil {
		c.Response.Status = 404
		return c.RenderText("Movie not found")
	}

	var movie models.Movie
	err = dynamodbattribute.UnmarshalMap(result.Item, &movie)
	if err != nil {
		revel.AppLog.Errorf("Failed to unmarshal movie: %v", err)
		c.Response.Status = 500
		return c.RenderText("Failed to parse movie data")
	}

	return c.RenderJSON(movie)
}

// POST /movies
func (c MovieController) CreateMovie() revel.Result {
	var movie models.Movie
	if err := c.Params.BindJSON(&movie); err != nil {
		c.Response.Status = 400
		return c.RenderText("Invalid input")
	}

	movie.ID = generateULID()

	av, err := dynamodbattribute.MarshalMap(movie)
	if err != nil {
		revel.AppLog.Errorf("Failed to marshal movie: %v", err)
		c.Response.Status = 500
		return c.RenderText("Failed to prepare movie data")
	}

	svc := app.GetDB()
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Movies"),
		Item:      av,
	})

	if err != nil {
		revel.AppLog.Errorf("Failed to insert movie: %v", err)
		c.Response.Status = 500
		return c.RenderText("Failed to create movie")
	}

	return c.RenderJSON(movie)
}

// PUT /movies/{movieId}
func (c MovieController) UpdateMovie(movieId string) revel.Result {
    var movie models.Movie
    if err := c.Params.BindJSON(&movie); err != nil {
        c.Response.Status = 400
        return c.RenderText("Invalid input")
    }

    svc := app.GetDB()

    // Create a map for updates to send to DynamoDB
    updates := map[string]*dynamodb.AttributeValueUpdate{}

    // Check if the fields are not empty and need to be updated
    if movie.Title != "" {
        updates["Title"] = &dynamodb.AttributeValueUpdate{
            Action: aws.String("PUT"),
            Value:  &dynamodb.AttributeValue{S: aws.String(movie.Title)},
        }
    }
    if movie.Plot != "" {
        updates["Plot"] = &dynamodb.AttributeValueUpdate{
            Action: aws.String("PUT"),
            Value:  &dynamodb.AttributeValue{S: aws.String(movie.Plot)},
        }
    }
    if movie.Rating != 0 { // Ensure Rating is an integer and not zero by default
        updates["Rating"] = &dynamodb.AttributeValueUpdate{
            Action: aws.String("PUT"),
            Value:  &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", movie.Rating))},
        }
    }
    if movie.Year != "" {
        updates["Year"] = &dynamodb.AttributeValueUpdate{
            Action: aws.String("PUT"),
            Value:  &dynamodb.AttributeValue{S: aws.String(movie.Year)},
        }
    }

    // Perform the update operation
    out, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
        TableName:        aws.String("Movies"),
        Key:              map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(movieId)}},
        AttributeUpdates: updates,
        ReturnValues:     aws.String("ALL_NEW"),
    })
    if err != nil {
        c.Response.Status = 500
        return c.RenderText(fmt.Sprintf("Failed to update movie: %v", err))
    }

    // Unmarshal the updated attributes into the Movie struct
    var updatedMovie models.Movie
    if err := dynamodbattribute.UnmarshalMap(out.Attributes, &updatedMovie); err != nil {
        c.Response.Status = 500
        return c.RenderText("Failed to unmarshal updated movie")
    }

    // Return the updated movie data as JSON
    return c.RenderJSON(updatedMovie)
}

// DELETE /movies/{movieId}
func (c MovieController) DeleteMovie(movieId string) revel.Result {
	svc := app.GetDB()

	// Check ng movie 
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String("Movies"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(movieId)},
		},
	}

	getResult, err := svc.GetItem(getInput)
	if err != nil {
		c.Response.Status = 500
		return c.RenderText("Failed to check movie existence")
	}

	if getResult.Item == nil {
		c.Response.Status = 404
		return c.RenderText("Movie not found")
	}

	// Proceed with deletion
	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Movies"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(movieId)},
		},
	})

	if err != nil {
		revel.AppLog.Errorf("Failed to delete movie: %v", err)
		c.Response.Status = 500
		return c.RenderText("Failed to delete movie")
	}

	return c.RenderText(fmt.Sprintf("Movie with ID %s deleted successfully.", movieId))
}
