package constants

import "fmt"

type OutreachType string
type ConnectionStatus string
type InMailStatus string
type LeadTemperature string
type ProfileType string

const (
	OutreachTypeConnection OutreachType = "connection"
	OutreachTypeInMail     OutreachType = "inMail"

	ConnectionStatusPending   ConnectionStatus = "pending"
	ConnectionStatusResponded ConnectionStatus = "responded"
	ConnectionStatusAccepted  ConnectionStatus = "accepted"

	LeadTemperatureCold LeadTemperature = "cold"
	LeadTemperatureHot  LeadTemperature = "hot"

	ProfileTypePublic  ProfileType = "public"
	ProfileTypePrivate ProfileType = "private"

	ErrorMessage string = "Something went wrong. Try again later."

	FormFieldKeyProfileType     string = "profileType"
	FormFieldKeyOutreachType    string = "outreachType"
	FormFieldKeyUrl             string = "url"
	FormFieldKeyName            string = "name"
	FormFieldKeyNotes           string = "notes"
	FormFieldConnectionStatus   string = "connectionStatus"
	FormFieldKeyLeadTemperature string = "leadTemperature"
	FormFieldFollowupSent       string = "followupSent"
	FormFieldPictureUrl string = "pictureUrl"
)

func ValidateOutReachType(value OutreachType) error {
	switch value {
	case OutreachTypeConnection, OutreachTypeInMail:
		return nil
	default:
		return fmt.Errorf("invalid outreach type: %s", value)
	}
}

func ValidateConnectionStatus(value ConnectionStatus) error {
	switch value {
	case ConnectionStatusPending, ConnectionStatusResponded, ConnectionStatusAccepted:
		return nil
	default:
		return fmt.Errorf("invalid connection type: %s", value)
	}
}

func ValidateLeadTemperature(value LeadTemperature) error {
	switch value {
	case LeadTemperatureHot, LeadTemperatureCold:
		return nil
	default:
		return fmt.Errorf("invalid lead temperature type: %s", value)
	}
}

func ValidateProfileType(value ProfileType) error {
	switch value {
	case ProfileTypePublic, ProfileTypePrivate:
		return nil
	default:
		return fmt.Errorf("invalid profile type: %s", value)
	}
}
