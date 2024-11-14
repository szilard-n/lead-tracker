package dto

import (
	"leadgentracker/internals/model/constants"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NewLeadProperties struct {
	ProfileType  constants.ProfileType
	OutreachType constants.OutreachType
	Url          string
	Name         string
	PictureUrl   string
}

type UpdateLeadProperties struct {
	ID               primitive.ObjectID
	ConnectionStatus constants.ConnectionStatus
	LeadTemperature  constants.LeadTemperature
	FollowupSent     bool
	Notes            string
}
