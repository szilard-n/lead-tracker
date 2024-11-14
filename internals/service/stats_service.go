package service

import (
	"context"
	"time"

	"leadgentracker/internals/model"
	"leadgentracker/internals/model/constants"
	"leadgentracker/internals/repository"
)

type StatsService struct {
	repo repository.StatsRepository
}

func NewStatsService(leadRepository repository.StatsRepository) *StatsService {
	return &StatsService{
		repo: leadRepository,
	}
}

func (s *StatsService) UpdateStats(ctx context.Context, outreachType constants.OutreachType) error {
	return s.repo.Update(ctx, outreachType)
}

func (s *StatsService) GetCurrentDayStats(ctx context.Context) (*model.Stats, error) {
	return s.repo.GetForDate(ctx, time.Now())
}

func (s *StatsService) GetTotalStats(ctx context.Context) (*model.Stats, error) {
	return s.repo.GetTotal(ctx)
}
