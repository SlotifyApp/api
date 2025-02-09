package scheduling

import (
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type Event struct {
	EventID       string
	StartDateTime time.Time
	EndDateTime   time.Time
	Location      string
	Attendees     []string
}

const (
	ConflictFree       = 0
	ConflictReschedule = 1
	ConflictTooMany    = 2
)

type Slot struct {
	StartDateTime     time.Time
	EndDateTime       time.Time
	Conflict          int // 0 = free, 1 = can reschedule, 2 = conflict
	RescheduleEventID string
}

func ParseEvents(eventable []models.Eventable) (map[string][]Event, error) {
	// Parse events
	allEvents := make(map[string][]Event)

	for _, event := range eventable {
		tempAttendees := []string{}

		for _, attendee := range event.GetAttendees() {
			tempAttendees = append(tempAttendees, *attendee.GetEmailAddress().GetAddress())
		}

		valStartTime, err := time.Parse(time.StampNano, *event.GetStart().GetDateTime())
		if err != nil {
			return nil, err
		}

		valEndTime, err := time.Parse(time.StampNano, *event.GetEnd().GetDateTime())
		if err != nil {
			return nil, err
		}

		currEvent := Event{
			EventID:       *event.GetId(),
			StartDateTime: valStartTime,
			EndDateTime:   valEndTime,
			Location:      *event.GetLocation().GetDisplayName(),
			Attendees:     tempAttendees,
		}

		key := currEvent.StartDateTime.Format("2006-01-02")
		allEvents[key] = append(allEvents[key], currEvent)
	}

	// Sort events per date by time

	return allEvents, nil
}

// nolint: funlen // Decrease function length
func ParseUserEventsForADay(slots []Slot, userEvents []Event, currDate time.Time) ([]Slot, error) {
	// Calculate slots for a day for a user
	startWorkingTime := currDate
	workingTime := 9
	endWorkingTime := currDate.Add(time.Hour * time.Duration(workingTime))

	// For each slot, and each event
	slotIndex := 0
	eventIndex := 0

	newSlots := []Slot{}

	//nolint: mnd // using magic number 15
	for i := startWorkingTime; i.Compare(endWorkingTime) != 1; i = i.Add(time.Minute * 15) {
		// Set currSlots and currEvent
		if slotIndex == len(slots) {
			return newSlots, nil
		}

		if eventIndex == len(userEvents) {
			// Add all remaining slots
			tempSlot := slots[slotIndex]
			if i.Compare(tempSlot.StartDateTime) != -1 {
				newSlots = append(newSlots, Slot{
					StartDateTime:     i,
					EndDateTime:       tempSlot.EndDateTime,
					Conflict:          tempSlot.Conflict,
					RescheduleEventID: tempSlot.RescheduleEventID,
				})
			}

			slotIndex++
			for ind := slotIndex; ind < len(slots); ind++ {
				newSlots = append(newSlots, slots[slotIndex])
			}

			return newSlots, nil
		}

		currSlot := slots[slotIndex]
		currEvent := userEvents[eventIndex]

		// Check where the slots and events are positioned/overlapped
		//nolint: nestif, gocritic // change to switch case
		if currSlot.EndDateTime.Compare(currEvent.StartDateTime) != 1 {
			// Slot is in front of the event
			newSlots = append(newSlots, currSlot)
			i = i.Add(currSlot.EndDateTime.Sub(currSlot.StartDateTime))

			// Choose the next slot
			slotIndex++
		} else if currSlot.StartDateTime.Compare(currEvent.EndDateTime) != -1 {
			// Slot if after the event
			i = currEvent.EndDateTime
			// Next event if there is one
			eventIndex++
		} else {
			// Trim excess slot space in front of event
			if currSlot.StartDateTime.Compare(currEvent.StartDateTime) == -1 {
				newSlots = append(newSlots, Slot{
					StartDateTime:     i,
					EndDateTime:       currEvent.StartDateTime,
					Conflict:          currSlot.Conflict,
					RescheduleEventID: currSlot.RescheduleEventID,
				})

				i = i.Add(currEvent.StartDateTime.Sub(currSlot.StartDateTime))
			}

			if i.Compare(currEvent.StartDateTime) == 0 {
				// the event and current slot position start at the same time

				// check which one finishes first
				if currSlot.EndDateTime.Compare(currEvent.EndDateTime) != 1 {
					// Increase conflict of slot
					newSlots = append(newSlots, Slot{
						StartDateTime:     i,
						EndDateTime:       currSlot.EndDateTime,
						Conflict:          currSlot.Conflict + 1,
						RescheduleEventID: currEvent.EventID,
					})

					// Slot has been successfully parsed
					i = currSlot.EndDateTime
					slotIndex++
				} else {
					// Increase conflict of slot
					newSlots = append(newSlots, Slot{
						StartDateTime:     i,
						EndDateTime:       currEvent.EndDateTime,
						Conflict:          currSlot.Conflict + 1,
						RescheduleEventID: currEvent.EventID,
					})

					// Slot partition is parsed correctly
					// Event is successfully finished
					i = currEvent.EndDateTime
					eventIndex++
				}
			}

			// Implement over lapping
		}
	}

	return newSlots, nil
}

func MergeDaySlots(currSlots []Slot) []Slot {
	// Merge slots to reduce multiple different slots
	mergedSlots := []Slot{}

	tempSlot := currSlots[0]

	for currSlotIndex := 1; currSlotIndex < len(currSlots); currSlotIndex++ {
		nextSlot := currSlots[currSlotIndex]

		// If the slots are the same types
		//nolint: nestif // use switchcase instead of nested if
		if tempSlot.Conflict == nextSlot.Conflict {
			// Check if they are both free slots
			//nolint: gocritic // user switchcase instead of nested if
			if tempSlot.Conflict == ConflictFree {
				tempSlot = Slot{
					StartDateTime:     tempSlot.StartDateTime,
					EndDateTime:       nextSlot.EndDateTime,
					Conflict:          ConflictFree,
					RescheduleEventID: "",
				}
			} else if tempSlot.Conflict == ConflictReschedule {
				// Check if they both are conflicted by the same event
				if tempSlot.RescheduleEventID == nextSlot.RescheduleEventID {
					tempSlot = Slot{
						StartDateTime:     tempSlot.StartDateTime,
						EndDateTime:       nextSlot.EndDateTime,
						Conflict:          ConflictReschedule,
						RescheduleEventID: tempSlot.RescheduleEventID,
					}
				} else {
					mergedSlots = append(mergedSlots, tempSlot)
					tempSlot = nextSlot
				}
			} else {
				tempSlot = Slot{
					StartDateTime:     tempSlot.StartDateTime,
					EndDateTime:       nextSlot.EndDateTime,
					Conflict:          ConflictTooMany,
					RescheduleEventID: "",
				}
			}
		}

		// If it is the last index
		if currSlotIndex == (len(currSlots) - 1) {
			if tempSlot.EndDateTime == nextSlot.EndDateTime {
				// Slots were merged
				mergedSlots = append(mergedSlots, tempSlot)
			} else {
				mergedSlots = append(mergedSlots, tempSlot)
				mergedSlots = append(mergedSlots, nextSlot)
			}
		}
	}

	return mergedSlots
}

func CalculateSlotScore(slot Slot) float64 {
	// Will also take in new event data
	// Will also call AI Model
	return float64(slot.Conflict)
}
