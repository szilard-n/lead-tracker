package repository

import (
	"context"
	"leadgentracker/internals/model"
	"leadgentracker/internals/model/constants"
	"leadgentracker/internals/model/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadRepository interface {
	Create(ctx context.Context, lead *model.Lead) error
	Update(ctx context.Context, updateProperties *dto.UpdateLeadProperties) (*model.Lead, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	ListPaged(ctx context.Context, filter *dto.LeadFilter) ([]model.Lead, int, error)
}

type StatsRepository interface {
	Update(cctx context.Context, outreach constants.OutreachType) error
	GetTotal(ctx context.Context) (*model.Stats, error)
	GetForDate(ctx context.Context, date time.Time) (*model.Stats, error)
}
