package repository

import (
	"context"
	"fmt"
	"leadgentracker/internals/model"
	"leadgentracker/internals/model/constants"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStatsRepository struct {
	db  *mongo.Client
	col *mongo.Collection
}

func NewStatsRepository(client *mongo.Client) *MongoStatsRepository {
	return &MongoStatsRepository{
		db:  client,
		col: client.Database(os.Getenv("MONGO_DB")).Collection("stats"),
	}
}

// Update method updates both totalStats and today's stats in dailyStats based on outreachType
func (r *MongoStatsRepository) Update(ctx context.Context, outreachType constants.OutreachType) error {
	// Prepare the increment field based on the outreach type
	var totalStatsField string
	var dailyStatsField string

	switch outreachType {
	case constants.OutreachTypeConnection:
		totalStatsField = "totalStats.connections"
		dailyStatsField = "dailyStats.$.connections"
	case constants.OutreachTypeInMail:
		totalStatsField = "totalStats.inMails"
		dailyStatsField = "dailyStats.$.inMails"
	default:
		return fmt.Errorf("unsupported outreach type: %v", outreachType)
	}

	// Update or insert Total Stats with upsert to handle first-time setup
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": "stats"},
		bson.M{
			"$inc": bson.M{
				totalStatsField: 1, // Increment only the relevant totalStats field
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to update total stats: %w", err)
	}

	today := time.Now().Format("2006-01-02") // format as YYYY-MM-DD

	// Attempt to update today's stats within dailyStats
	updateResult, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": "stats", "dailyStats.date": today},
		bson.M{
			"$inc": bson.M{
				dailyStatsField: 1, // Increment only the relevant dailyStats field
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update today's stats: %w", err)
	}

	// If no document was modified, add today's date as a new entry in dailyStats
	if updateResult.ModifiedCount == 0 {
		newDailyStats := bson.M{
			"date":        today,
			"connections": 0,
			"inMails":     0,
		}

		// Set the initial count based on the outreach type
		if outreachType == constants.OutreachTypeConnection {
			newDailyStats["connections"] = 1
		} else if outreachType == constants.OutreachTypeInMail {
			newDailyStats["inMails"] = 1
		}

		_, err = r.col.UpdateOne(
			ctx,
			bson.M{"_id": "stats"},
			bson.M{
				"$push": bson.M{
					"dailyStats": newDailyStats,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to insert new daily stats for today: %w", err)
		}
	}

	return nil
}

// GetTotal retrieves the totalStats from the stats document
func (r *MongoStatsRepository) GetTotal(ctx context.Context) (*model.Stats, error) {
	var result struct {
		TotalStats model.Stats `bson:"totalStats"`
	}

	err := r.col.FindOne(ctx, bson.M{"_id": "stats"}).Decode(&result)
	if err != nil {
		// Return a zero-value Stats object if no document is found
		if err == mongo.ErrNoDocuments {
			return &model.Stats{Connections: 0, InMails: 0}, nil
		}
		// Return the actual error if it's not due to missing document
		return nil, fmt.Errorf("failed to retrieve total stats: %w", err)
	}

	return &result.TotalStats, nil
}

// GetForDate retrieves stats for a specific date from the dailyStats array
func (r *MongoStatsRepository) GetForDate(ctx context.Context, date time.Time) (*model.Stats, error) {
	formattedDate := date.Format("2006-01-02") // YYYY-MM-DD

	// Use an aggregation pipeline to filter by date in the array
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": "stats"}}},
		{{Key: "$project", Value: bson.M{
			"dailyStats": bson.M{
				"$filter": bson.M{
					"input": "$dailyStats",
					"as":    "day",
					"cond":  bson.M{"$eq": []interface{}{"$$day.date", formattedDate}},
				},
			},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stats for date: %w", err)
	}
	defer cursor.Close(ctx)

	// Check if we have results in the cursor
	if cursor.Next(ctx) {
		var result struct {
			DailyStats []model.Stats `bson:"dailyStats"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode daily stats: %w", err)
		}
		// If no matching daily stats entry is found for the date, return zero-value Stats
		if len(result.DailyStats) > 0 {
			return &result.DailyStats[0], nil
		}
	}

	// Return a zero-value Stats object if no daily stats entry is found for the date
	return &model.Stats{Connections: 0, InMails: 0}, nil
}
