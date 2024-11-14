package service

import (
	"context"
	"time"

	"leadgentracker/internals/model"
	"leadgentracker/internals/model/constants"
	"leadgentracker/internals/model/dto"
	"leadgentracker/internals/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadService struct {
	repo repository.LeadRepository
}

func NewLeadService(repository repository.LeadRepository) *LeadService {
	return &LeadService{
		repo: repository,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, leadProperties *dto.NewLeadProperties) error {
	return s.repo.Create(ctx, &model.Lead{
		ID:               primitive.NewObjectID(),
		ConnectionStatus: constants.ConnectionStatusPending,
		LeadTemperature:  constants.LeadTemperatureCold,
		ProfileType:      leadProperties.ProfileType,
		OutreachType:     leadProperties.OutreachType,
		Date:             time.Now(),
		URL:              leadProperties.Url,
		Name:             leadProperties.Name,
		FollowupSent:     false,
		Notes:            "",
		PictureUrl:       leadProperties.PictureUrl,
	})
}

func (s *LeadService) UpdateLead(ctx context.Context, updatedLead *dto.UpdateLeadProperties) (*model.Lead, error) {
	return s.repo.Update(ctx, updatedLead)
}

func (s *LeadService) DeleteLead(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *LeadService) GetAllLeadsPaged(ctx context.Context, filter *dto.LeadFilter) ([]model.Lead, int, error) {
	return s.repo.ListPaged(ctx, filter)
}
