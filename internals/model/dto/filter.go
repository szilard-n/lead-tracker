package dto

import (
	"fmt"
	"leadgentracker/internals/model/constants"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const LeadsPerPage = 6

type LeadFilter struct {
	SearchQuery     string
	OutreachType    constants.OutreachType
	LeadTemperature constants.LeadTemperature
	DateAdded       time.Time
	Page            int
	LeadsPerPage    int
}

func (f LeadFilter) HasActiveFilters() bool {
	return f.SearchQuery != "" ||
		f.OutreachType != constants.OutreachType("") ||
		f.LeadTemperature != constants.LeadTemperature("") ||
		!f.DateAdded.IsZero()
}

func NewPagedLeadFilter(page int) *LeadFilter {
	return &LeadFilter{Page: page, LeadsPerPage: LeadsPerPage}
}

func NewLeadFilter(urlValues url.Values) (*LeadFilter, error) {
	filter := &LeadFilter{LeadsPerPage: LeadsPerPage}
	var errs []string

	// Extract and validate search query
	filter.SearchQuery = urlValues.Get("search")

	// Extract and validate outreach type
	if outreachType := urlValues.Get("outreachType"); outreachType != "" {
		switch constants.OutreachType(outreachType) {
		case constants.OutreachTypeConnection, constants.OutreachTypeInMail:
			filter.OutreachType = constants.OutreachType(outreachType)
		default:
			errs = append(errs, fmt.Sprintf("invalid outreach type: %s", outreachType))
		}
	}

	// Extract and validate lead temperature
	if leadTemp := urlValues.Get("leadTemperature"); leadTemp != "" {
		switch constants.LeadTemperature(leadTemp) {
		case constants.LeadTemperatureHot, constants.LeadTemperatureCold:
			filter.LeadTemperature = constants.LeadTemperature(leadTemp)
		default:
			errs = append(errs, fmt.Sprintf("invalid lead temperature: %s", leadTemp))
		}
	}

	// Extract and validate date
	if dateStr := urlValues.Get("dateAdded"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid date format: %s", dateStr))
		} else {
			// Ensure date is not in the future
			if date.After(time.Now()) {
				errs = append(errs, "date cannot be in the future")
			} else {
				filter.DateAdded = date
			}
		}
	}

	// Handle page number
	if pageStr := urlValues.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			errs = append(errs, "invalid page number")
		} else if page < 1 {
			errs = append(errs, "page number must be positive")
		} else {
			filter.Page = page
		}
	} else {
		filter.Page = 1 // Default to first page
	}

	// Return any validation errors
	var err error
	if len(errs) > 0 {
		err = fmt.Errorf("filter validation errors: %s", strings.Join(errs, "; "))
	}

	return filter, err
}
