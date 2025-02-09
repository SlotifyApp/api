package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlotifyApp/slotify-backend/jwt"
	"github.com/SlotifyApp/slotify-backend/scheduling"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"go.uber.org/zap"
)

// (POST /api/algo).
// nolint: funlen, gocognit // Decrease function length
func (s Server) PostAPIAlgo(w http.ResponseWriter, r *http.Request) {
	var teamBody PostAPIAlgoJSONRequestBody
	var err error

	// Parse request body for errors
	if err = json.NewDecoder(r.Body).Decode(&teamBody); err != nil {
		s.Logger.Error(ErrUnmarshalBody, zap.Error(err))
		sendError(w, http.StatusBadRequest, ErrUnmarshalBody.Error())
		return
	}

	// Create Microsft Graph ClientS
	userID, err := jwt.GetUserIDFromReq(r)
	if err != nil {
		s.Logger.Error("failed to get userid from request access token")
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	at, err := getMSFTAccessToken(context.Background(), s.MSALClient, s.DB, userID)
	if err != nil {
		s.Logger.Error("failed to get microsoft access token", zap.Error(err))
		sendError(w, http.StatusUnauthorized, "Try again later.")
		return
	}

	graph, err := createMSFTGraphClient(at)

	if err != nil || graph == nil {
		s.Logger.Error("failed to create msft graph client", zap.Error(err))
		sendError(w, http.StatusInternalServerError, "msft graph could not be created")
		return
	}

	//	Get all user's meeting data
	type UserEvents struct {
		userID string
		events map[string][]scheduling.Event
	}

	userEvents := []UserEvents{}
	participants := *teamBody.Participants // array of emails

	for _, participantEmail := range participants {
		// Get Participant's calendar events, (sort by date in the future)
		temp := string(participantEmail)

		var events models.EventCollectionResponseable
		events, err = graph.Users().ByUserId(temp).Calendar().Events().Get(context.Background(), nil)
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

	const startWorkingTime = 8
	const endWorkingTime = 17
	const workingHours = endWorkingTime - startWorkingTime

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
			if int(eachSlot.EndDateTime.Sub(eachSlot.StartDateTime)) >= *teamBody.EventDuration {
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
