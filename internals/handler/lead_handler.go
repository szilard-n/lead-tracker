package handler

import (
	"log"
	"net/http"
	"strconv"

	"leadgentracker/internals/model/constants"
	"leadgentracker/internals/model/dto"
	"leadgentracker/internals/service"
	"leadgentracker/views"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MsgLeadUpdateSuccess     = "Lead updated successfully!"
	MsgLeadUpdateError       = "Failed to update lead. Please try again."
	MsgLeadDeleteSuccess     = "Lead deleted successfully!"
	MsgLeadDeleteError       = "Failed to delete lead. Please try again."
	MsgLeadListFilterWarning = "Invalid filters provided. Please try again."
	MsgLeadListError         = "Failed to fetch leads. Please try again."
)

type LeadHandler struct {
	ls *service.LeadService
	ss *service.StatsService
	b  *SSEBroadcaster
}

func NewLeadHandler(leadService *service.LeadService, statsService *service.StatsService, sseBroadcaster *SSEBroadcaster) *LeadHandler {
	return &LeadHandler{
		ls: leadService,
		ss: statsService,
		b:  sseBroadcaster,
	}
}

func (h *LeadHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("serving index file")
	totalStats, err := h.ss.GetTotalStats(r.Context())
	if err != nil {
		log.Printf("failed to fetch total stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	todayStats, err := h.ss.GetCurrentDayStats(r.Context())
	if err != nil {
		log.Printf("failed to fetch today's stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	filter := dto.NewPagedLeadFilter(1)

	leads, totalPages, err := h.ls.GetAllLeadsPaged(r.Context(), filter)
	if err != nil {
		log.Printf("failed to fetch leads: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	log.Printf("Fetched %d leands and total pages %d", len(leads), totalPages)

	// Render the Index page with data
	if err := views.Index(totalStats, todayStats, leads, totalPages, filter).Render(r.Context(), w); err != nil {
		log.Printf("failed to render index page: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}
}

func (h *LeadHandler) AddLead(w http.ResponseWriter, r *http.Request) {
	log.Println("adding new lead")
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to parse form: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
		return
	}

	profileType := constants.ProfileType(r.FormValue(constants.FormFieldKeyProfileType))
	outreachType := constants.OutreachType(r.FormValue(constants.FormFieldKeyOutreachType))

	if constants.ValidateProfileType(profileType) != nil || constants.ValidateOutReachType(outreachType) != nil {
		log.Printf("invalid fields provided for profile and outereach types: %s", r.Form.Encode())
		http.Error(w, "invalid fields provided", http.StatusBadRequest)
		return
	}

	err := h.ls.CreateLead(r.Context(), &dto.NewLeadProperties{
		ProfileType:  profileType,
		OutreachType: outreachType,
		Url:          r.FormValue(constants.FormFieldKeyUrl),
		Name:         r.FormValue(constants.FormFieldKeyName),
		PictureUrl:   r.FormValue(constants.FormFieldPictureUrl),
	})
	if err != nil {
		log.Printf("failed to create new lead: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	if err := h.ss.UpdateStats(r.Context(), outreachType); err != nil {
		log.Printf("failed to update stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	// Set HTMX triggers to refresh lead stats and lead list
	h.b.Broadcast("refreshLeadList,refreshLeadStats,renderNewLeadNotification")
	w.WriteHeader(http.StatusOK)
}

func (h *LeadHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("failed to parse form: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		log.Printf("missing ID in URL: %s", r.URL.Query().Encode())
		http.Error(w, "missing lead ID", http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		return
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("invalid hex ID provided: %s: %s", id, err)
		http.Error(w, "invalid ID provided", http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		return
	}

	connectionStatus := constants.ConnectionStatus(r.FormValue(constants.FormFieldConnectionStatus))
	leadTemperature := constants.LeadTemperature(r.FormValue(constants.FormFieldKeyLeadTemperature))

	// Create lead update from form values
	updateProps := &dto.UpdateLeadProperties{
		ID:               objectId,
		ConnectionStatus: connectionStatus,
		LeadTemperature:  leadTemperature,
		FollowupSent:     r.FormValue(constants.FormFieldFollowupSent) != "",
		Notes:            r.FormValue(constants.FormFieldKeyNotes),
	}

	// Update the lead
	updatedLead, err := h.ls.UpdateLead(r.Context(), updateProps)
	if err != nil {
		log.Printf("failed to update lead: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			log.Printf("error while updating lead: invalid page number: %s [%v]", pageStr, err)
			http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
			h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		}
		page = parsedPage
	}

	// Render the updated lead details
	if err := views.Lead(updatedLead, page).Render(r.Context(), w); err != nil {
		log.Printf("failed to render lead: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		h.renderNotification(w, r, views.NotificationError, MsgLeadUpdateError)
		return
	}
	h.renderNotification(w, r, views.NotificationSuccess, MsgLeadUpdateSuccess)
}

func (h *LeadHandler) DeleteLead(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		log.Printf("missing ID in URL: %s", r.URL.Query().Encode())
		http.Error(w, "missing lead ID", http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
		return
	}

	log.Printf("deleting lead with ID: %s", id)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("invalid hex ID provided: %s: %s", id, err)
		http.Error(w, "invalid ID provided", http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
		return
	}

	err = h.ls.DeleteLead(r.Context(), objectId)
	if err != nil {
		log.Printf("error while deleting lead: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			log.Printf("invalid page number provided: %s [%v]", pageStr, err)
			http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
			h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
			return
		}
		page = parsedPage
	}

	filter := dto.NewPagedLeadFilter(page)
	leads, totalPages, err := h.ls.GetAllLeadsPaged(r.Context(), filter)
	if err != nil {
		log.Printf("error while deleting lead: failed to fetch remaining leads: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
		h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
	}

	// if only one page left, go to first page
	if totalPages == 1 {
		filter.Page = 1
		leads, totalPages, err = h.ls.GetAllLeadsPaged(r.Context(), filter)
		if err != nil {
			log.Printf("error while deleting lead: failed to fetch remaining leads: %s", err)
			http.Error(w, constants.ErrorMessage, http.StatusBadRequest)
			h.renderNotification(w, r, views.NotificationError, MsgLeadDeleteError)
		}
	}

	if err := views.LeadList(leads, totalPages, filter).Render(r.Context(), w); err != nil {
		log.Printf("failed to render lead list: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		h.renderNotification(w, r, views.NotificationWarning, MsgLeadListError)
		return
	}
	h.renderNotification(w, r, views.NotificationSuccess, MsgLeadDeleteSuccess)
}

func (h *LeadHandler) GetAllLeads(w http.ResponseWriter, r *http.Request) {
	log.Println("Received get all leads request")

	filter, err := dto.NewLeadFilter(r.URL.Query())
	if err != nil {
		log.Printf("[WARNING] Invalid filter values provided: %s", err)
		h.renderNotification(w, r, views.NotificationWarning, MsgLeadListFilterWarning)
	}

	leads, totalPages, err := h.ls.GetAllLeadsPaged(r.Context(), filter)
	if err != nil {
		log.Printf("failed to fetch leads: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		h.renderNotification(w, r, views.NotificationWarning, MsgLeadListError)
		return
	}

	if err := views.LeadList(leads, totalPages, filter).Render(r.Context(), w); err != nil {
		log.Printf("failed to render lead list: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		h.renderNotification(w, r, views.NotificationWarning, MsgLeadListError)
		return
	}
}

func (h *LeadHandler) GetLeadStats(w http.ResponseWriter, r *http.Request) {
	log.Println("getting all lead stats")
	totalStats, err := h.ss.GetTotalStats(r.Context())
	if err != nil {
		log.Printf("failed to fetch total stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	todayStats, err := h.ss.GetCurrentDayStats(r.Context())
	if err != nil {
		log.Printf("failed to fetch today's stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	if err := views.LeadStats(totalStats, todayStats).Render(r.Context(), w); err != nil {
		log.Printf("failed to render lead stats: %s", err)
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}
}

func (h *LeadHandler) renderNotification(w http.ResponseWriter, r *http.Request, nType views.NotificationType, message string) {
	notification := views.NotificationProps{
		Type:    nType,
		Message: message,
	}

	w.Header().Set("HX-Swap", "innerHTML") // Ensures the new notification replaces any existing ones
	if err := views.Notification(notification).Render(r.Context(), w); err != nil {
		log.Printf("[ERROR] Failed to render notification: %v", err)
	}
}
