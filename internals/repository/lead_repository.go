package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"leadgentracker/internals/model"
	"leadgentracker/internals/model/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoFieldID               = "_id"
	MongoFieldName             = "name"
	MongoFieldOutreachType     = "outreachtype"
	MongoFieldLeadTemp         = "leadtemperature"
	MongoFieldConnectionStatus = "connectionstatus"
	MongoFieldFollowupSent     = "followupsent"
	MongoFieldDate             = "date"
	MongoFieldProfileType      = "profiletype"
	MongoFieldURL              = "url"
	MongoFieldPictureURL       = "pictureurl"
	MongoFieldNotes            = "notes"
)

type MongoLeadRepository struct {
	db  *mongo.Client
	col *mongo.Collection
}

func NewLeadRepository(client *mongo.Client) *MongoLeadRepository {
	return &MongoLeadRepository{
		db:  client,
		col: client.Database(os.Getenv("MONGO_DB")).Collection("leads"),
	}
}

func (r *MongoLeadRepository) Create(ctx context.Context, lead *model.Lead) error {
	_, err := r.col.InsertOne(ctx, lead)
	if err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}
	return nil
}

func (r *MongoLeadRepository) Update(ctx context.Context, updateProperties *dto.UpdateLeadProperties) (*model.Lead, error) {
	// Filter for finding the lead by ID
	filter := bson.D{{Key: "_id", Value: updateProperties.ID}}

	// Construct the update document directly, since validation is handled beforehand
	update := bson.D{
		{Key: MongoFieldConnectionStatus, Value: updateProperties.ConnectionStatus},
		{Key: MongoFieldLeadTemp, Value: updateProperties.LeadTemperature},
		{Key: MongoFieldFollowupSent, Value: updateProperties.FollowupSent},
		{Key: MongoFieldNotes, Value: updateProperties.Notes},
	}

	// Use FindOneAndUpdate to perform the update and retrieve the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedLead model.Lead
	err := r.col.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: update}}, opts).Decode(&updatedLead)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("lead with ID %s not found", updateProperties.ID.Hex())
		}
		return nil, fmt.Errorf("failed to update lead: %w", err)
	}

	return &updatedLead, nil
}

func (r *MongoLeadRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.D{{Key: MongoFieldID, Value: id}}
	_, err := r.col.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}
	return nil
}

func (r *MongoLeadRepository) ListPaged(ctx context.Context, filter *dto.LeadFilter) ([]model.Lead, int, error) {
	var leads []model.Lead

	// Build query filters
	query := r.buildFilters(filter)

	// Count total matching documents (with filters applied)
	totalCount, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leads: %w", err)
	}

	// Calculate pagination values
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.LeadsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	skip := (filter.Page - 1) * filter.LeadsPerPage

	// Set up find options with pagination and sorting
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(filter.LeadsPerPage)).
		SetSort(bson.D{{Key: MongoFieldDate, Value: -1}})

	// Execute query with filters and options
	cursor, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, fmt.Errorf("failed to decode leads: %w", err)
	}

	return leads, totalPages, nil
}

// buildFilters constructs the MongoDB query filter based on the provided LeadFilter
func (r *MongoLeadRepository) buildFilters(filter *dto.LeadFilter) bson.D {
	filters := bson.A{}

	// Add name search filter if provided
	if filter.SearchQuery != "" {
		filters = append(filters, bson.D{{
			Key: MongoFieldName,
			Value: primitive.Regex{
				Pattern: filter.SearchQuery,
				Options: "i",
			},
		}})
	}

	// Add outreach type filter if provided
	if filter.OutreachType != "" {
		outreachFilter := bson.D{{
			Key:   MongoFieldOutreachType,
			Value: filter.OutreachType,
		}}
		filters = append(filters, outreachFilter)
	}

	// Add lead temperature filter if provided
	if filter.LeadTemperature != "" {
		tempFilter := bson.D{{
			Key:   MongoFieldLeadTemp,
			Value: filter.LeadTemperature,
		}}
		filters = append(filters, tempFilter)
	}

	// Add date filter if provided
	if !filter.DateAdded.IsZero() {
		startOfDay := time.Date(
			filter.DateAdded.Year(),
			filter.DateAdded.Month(),
			filter.DateAdded.Day(),
			0, 0, 0, 0,
			filter.DateAdded.Location(),
		)
		endOfDay := startOfDay.Add(24 * time.Hour)

		filters = append(filters, bson.D{{
			Key: MongoFieldDate,
			Value: bson.D{
				{Key: "$gte", Value: startOfDay},
				{Key: "$lt", Value: endOfDay},
			},
		}})
	}

	finalFilter := bson.D{}
	if len(filters) > 0 {
		finalFilter = bson.D{{
			Key:   "$and",
			Value: filters,
		}}
	}
	return finalFilter
}
