package controllers

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/revel/revel"

	"MoviesApp/app"
	"MoviesApp/app/models"
)

type MovieController struct {
	*revel.Controller
}

// Helper function to generate sequential IDs like 0001, 0002, 0003, etc.
func GenerateSequentialID() string {
	svc := app.GetDB()

	// Fetch the current counter value from DynamoDB
	counterParams := &dynamodb.GetItemInput{
		TableName: aws.String("Counters"), // Assuming you have a table named 'Counters'
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String("counter")},
		},
	}

	result, err := svc.GetItem(counterParams)
	if err != nil {
		revel.AppLog.Errorf("Failed to fetch counter: %v", err)
		return ""
	}

	// If counter item doesn't exist, initialize it with LastUsedID = 0
	counterValue := 0
	if result.Item == nil {
		// Initialize counter value in DynamoDB
		_, err := svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("Counters"),
			Item: map[string]*dynamodb.AttributeValue{
				"ID":         {S: aws.String("counter")},
				"LastUsedID": {N: aws.String("0")},
			},
		})
		if err != nil {
			revel.AppLog.Errorf("Failed to initialize counter: %v", err)
			return ""
		}
	} else {
		// Retrieve and increment the counter value
		if val, err := strconv.Atoi(*result.Item["LastUsedID"].N); err == nil {
			counterValue = val
		}
	}

	// Increment the counter value
	counterValue++

	// Update the counter value in DynamoDB
	_, err = svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Counters"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String("counter")},
		},
		AttributeUpdates: map[string]*dynamodb.AttributeValueUpdate{
			"LastUsedID": {
				Action: aws.String("PUT"),
				Value:  &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", counterValue))},
			},
		},
	})

	if err != nil {
		revel.AppLog.Errorf("Failed to update counter: %v", err)
		return ""
	}

	// Return the formatted sequential ID
	return fmt.Sprintf("%04d", counterValue)
}

// GET /movies
func (c MovieController) ListMovies() revel.Result {
	svc := app.GetDB()
	params := &dynamodb.ScanInput{
		TableName: aws.String("Movies"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		revel.AppLog.Errorf("Failed to fetch movies: %v", err)
		c.Response.Status = 500
		return c.RenderText(fmt.Sprintf("Failed to fetch movies: %v", err))
	}

	movies := []models.Movie{}
	for _, item := range result.Items {
		movie := models.Movie{
			ID:    *item["ID"].S,
			Title: *item["Title"].S,
			Plot:  *item["Plot"].S,
			Year:  *item["Year"].S,
		}
		if rating, err := strconv.ParseFloat(*item["Rating"].N, 64); err == nil {
			movie.Rating = rating
		}
		movies = append(movies, movie)
	}

	revel.AppLog.Infof("Fetched movies: %+v", movies)
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
		return c.RenderText(fmt.Sprintf("Failed to fetch movie: %v", err))
	}

	if result.Item == nil {
		c.Response.Status = 404
		return c.RenderText("Movie not found")
	}

	movie := models.Movie{
		ID:    *result.Item["ID"].S,
		Title: *result.Item["Title"].S,
		Plot:  *result.Item["Plot"].S,
		Year:  *result.Item["Year"].S,
	}
	if rating, err := strconv.ParseFloat(*result.Item["Rating"].N, 64); err == nil {
		movie.Rating = rating
	}

	revel.AppLog.Infof("Fetched movie: %+v", movie)
	return c.RenderJSON(movie)
}

// POST /movies
func (c MovieController) CreateMovie() revel.Result {
	var movie models.Movie
	if err := c.Params.BindJSON(&movie); err != nil {
		revel.AppLog.Errorf("Error binding JSON: %v", err)
		c.Response.Status = 400
		return c.RenderText("Invalid input")
	}

	// Generate a sequential ID
	movie.ID = GenerateSequentialID()
	if movie.ID == "" {
		c.Response.Status = 500
		return c.RenderText("Failed to generate movie ID")
	}

	// Log the movie details
	revel.AppLog.Infof("Movie created from request: %+v", movie)

	svc := app.GetDB()
	_, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Movies"),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":     {S: aws.String(movie.ID)},
			"Title":  {S: aws.String(movie.Title)},
			"Plot":   {S: aws.String(movie.Plot)},
			"Rating": {N: aws.String(fmt.Sprintf("%.1f", movie.Rating))},
			"Year":   {S: aws.String(movie.Year)},
		},
	})

	if err != nil {
		revel.AppLog.Errorf("Error inserting into DynamoDB: %v", err)
		c.Response.Status = 500
		return c.RenderText(fmt.Sprintf("Failed to create movie: %v", err))
	}

	revel.AppLog.Infof("Movie successfully created: %+v", movie)
	return c.RenderJSON(movie)
}

// PUT /movies/{movieId}
func (c MovieController) UpdateMovie(movieId string) revel.Result {
	var movie models.Movie
	if err := c.Params.BindJSON(&movie); err != nil {
		revel.AppLog.Errorf("Error binding JSON: %v", err)
		c.Response.Status = 400
		return c.RenderText("Invalid input")
	}

	revel.AppLog.Infof("Movie update request: %+v", movie)

	svc := app.GetDB()
	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Movies"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(movieId)},
		},
		AttributeUpdates: map[string]*dynamodb.AttributeValueUpdate{
			"Title": {
				Action: aws.String("PUT"),
				Value:  &dynamodb.AttributeValue{S: aws.String(movie.Title)},
			},
			"Plot": {
				Action: aws.String("PUT"),
				Value:  &dynamodb.AttributeValue{S: aws.String(movie.Plot)},
			},
			"Rating": {
				Action: aws.String("PUT"),
				Value:  &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%.1f", movie.Rating))},
			},
			"Year": {
				Action: aws.String("PUT"),
				Value:  &dynamodb.AttributeValue{S: aws.String(movie.Year)},
			},
		},
	})

	if err != nil {
		revel.AppLog.Errorf("Error updating movie in DynamoDB: %v", err)
		c.Response.Status = 500
		return c.RenderText(fmt.Sprintf("Failed to update movie: %v", err))
	}

	revel.AppLog.Infof("Movie updated successfully: %+v", movie)
	return c.RenderJSON(movie)
}

// DELETE /movies/{movieId}
func (c MovieController) DeleteMovie(movieId string) revel.Result {
	// Log incoming request for better visibility
	revel.AppLog.Infof("Attempting to delete movie with ID: %s", movieId)

	// Initialize DynamoDB service
	svc := app.GetDB()

	// Prepare the delete request for DynamoDB
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String("Movies"),  // Ensure this is your correct table name
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(movieId)}, // Ensure the key matches the partition key in your DynamoDB table
		},
	}

	// Attempt to delete the item from DynamoDB
	_, err := svc.DeleteItem(params)
	if err != nil {
		// Log the error for better insight into the issue
		revel.AppLog.Errorf("Error while attempting to delete movie with ID %s: %v", movieId, err)
		c.Response.Status = 500
		return c.RenderText(fmt.Sprintf("Failed to delete movie with ID %s: %v", movieId, err))
	}

	// Log successful deletion
	revel.AppLog.Infof("Successfully deleted movie with ID: %s", movieId)

	// Return success message
	return c.RenderText(fmt.Sprintf("Movie with ID %s deleted successfully.", movieId))
}

