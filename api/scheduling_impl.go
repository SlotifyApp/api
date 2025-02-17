package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/scheduling"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

const (
	startWorkingTime = 8
	endWorkingTime   = 17
	workingHours     = endWorkingTime - startWorkingTime
)

// (POST /api/scheduling/free)
// nolint: funlen, gocognit // Decrease function length
func (s Server) PostAPISchedulingFree(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	var schedulingBody PostAPISchedulingFreeJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&schedulingBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	userID, ok := r.Context().Value(UserIDCtxKey{}).(uint32)
	if !ok {
		s.Logger.Error("failed to get userid from request context")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := CreateMSFTGraphClient(ctx, s.MSALClient, s.DB, userID)

	if err != nil || graph == nil {
		s.Logger.Error("failed to create msft graph client", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "msft graph could not be created")
		return
	}

	// Get all user's meeting data
	type UserEvents struct {
		userID string
		events map[string][]scheduling.Event
	}

	userEvents := []UserEvents{}

	for _, participantEmail := range schedulingBody.Participants {
		// Get Participant's calendar events, (sort by date in the future)
		temp := string(participantEmail)

		var events models.EventCollectionResponseable
		events, err = graph.Users().ByUserId(temp).Calendar().Events().Get(ctx, nil)
		if err != nil {
			s.Logger.Error("failed to call graph client route", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "failed to call graph client route")
			return
		}

		var allEvents map[string][]scheduling.Event
		allEvents, err = scheduling.ParseEvents(events.GetValue())
		if err != nil {
			s.Logger.Error("Failed to parse events", zap.Error(err))
			sendError(w, http.StatusInternalServerError, "failed to parse events for a specific user")
			return
		}

		// Append all the events of the users
		userEvents = append(userEvents, UserEvents{
			userID: temp,
			events: allEvents,
		})
	}

	//	for each user, calculate free slots
	currSlots := make(map[string][]scheduling.Slot)

	for _, users := range userEvents {
		// Assume that users.events is sorted by date

		for dayKey, day := range users.events {
			// Calculate free slots with conflicts and append/change
			var currDate time.Time
			currDate, err = time.Parse("2006-01-02", dayKey)
			if err != nil {
				s.Logger.Error("failed to parse date", zap.Error(err))
				sendError(w, http.StatusInternalServerError, "failed to parse date")
				return
			}

			currDate = currDate.Add(time.Hour * startWorkingTime)

			// check if map contains key
			_, exists := currSlots[dayKey]

			if !exists {
				currSlots[dayKey] = []scheduling.Slot{}
				currSlots[dayKey] = append(currSlots[dayKey], scheduling.Slot{
					StartDateTime:     currDate,
					EndDateTime:       currDate.Add(time.Hour * workingHours),
					Conflict:          0,
					RescheduleEventID: "",
				})
			}

			// Call parse events
			var newSlots []scheduling.Slot
			newSlots, err = scheduling.ParseUserEventsForADay(currSlots[dayKey], day, currDate)
			if err != nil {
				s.Logger.Error("failed to parse user slots for a user for a specific day", zap.Error(err))
				sendError(w, http.StatusInternalServerError, "failed to parse user slots for a user for a specific day")
				return
			}

			currSlots[dayKey] = newSlots
		}
	}

	// Find all slots where the new event can fit in both array

	type FinalSlot struct {
		StartDateTime       time.Time
		EndDateTime         time.Time
		Conflict            int
		Rating              float64
		ReschedulingEventID string
	}

	finalSlotsWithRatings := make(map[string][]FinalSlot)

	// Merge all slots
	mergedSlots := make(map[string][]scheduling.Slot)

	for dayKey, day := range currSlots {
		mergedSlots[dayKey] = scheduling.MergeDaySlots(day)
	}

	// Check for valid slots and calculate rating
	// Need to change to consider multuiple slots so an event can be in two slots
	for dayKey, slots := range mergedSlots {
		finalSlots := []scheduling.Slot{}

		// For eachslot, check if it can fit the new event
		for _, eachSlot := range slots {
			if eachSlot.EndDateTime.Sub(eachSlot.StartDateTime).Minutes() >= float64(schedulingBody.EventDuration) {
				// Check if slot is in currSlotsWithConflicts
				finalSlots = append(finalSlots, eachSlot)
			}
		}

		// For each valid slot
		// Calculate score for each valid slot
		for _, vSlot := range finalSlots {
			finalSlotsWithRatings[dayKey] = append(finalSlotsWithRatings[dayKey], FinalSlot{
				StartDateTime:       vSlot.StartDateTime,
				EndDateTime:         vSlot.EndDateTime,
				Conflict:            vSlot.Conflict,
				Rating:              scheduling.CalculateSlotScore(vSlot),
				ReschedulingEventID: vSlot.RescheduleEventID,
			})
		}
	}

	SetHeaderAndWriteResponse(w, http.StatusOK, finalSlotsWithRatings)
}
