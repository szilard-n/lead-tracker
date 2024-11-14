package model

import (
	"time"

	"leadgentracker/internals/model/constants"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Lead struct {
	ID               primitive.ObjectID         `bson:"_id,ommitempty"`
	ConnectionStatus constants.ConnectionStatus `json:"connectionStatus"`
	LeadTemperature  constants.LeadTemperature  `json:"leadTemperature"`
	ProfileType      constants.ProfileType      `json:"profileType"`
	OutreachType     constants.OutreachType     `json:"outreachType"`
	Date             time.Time                  `json:"date"`
	URL              string                     `json:"url"`
	Name             string                     `json:"name"`
	FollowupSent     bool                       `json:"followupSent"`
	Notes            string                     `json:"notes"`
	PictureUrl       string                     `json:"pictureUrl"`
}
