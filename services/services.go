package services

import (
	"errors"
	"sort"
	"time"

	"github.com/krushnna/meeting-scheduler/models"
	"github.com/krushnna/meeting-scheduler/repository"
)

// EventService handles business logic for events
type EventService struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) CreateEvent(event *models.Event) error {
	if event.Title == "" {
		return errors.New("event title is required")
	}
	if event.DurationMinutes <= 0 {
		return errors.New("event duration must be positive")
	}
	return s.repo.Create(event)
}

func (s *EventService) GetEvent(id uint) (*models.Event, error) {
	return s.repo.FindByID(id)
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	return s.repo.FindAll()
}

func (s *EventService) UpdateEvent(id uint, event *models.Event) error {
	if event.Title == "" {
		return errors.New("event title is required")
	}
	if event.DurationMinutes <= 0 {
		return errors.New("event duration must be positive")
	}
	return s.repo.Update(id, event)
}

func (s *EventService) DeleteEvent(id uint) error {
	return s.repo.Delete(id)
}

// TimeSlotService handles business logic for time slots
type TimeSlotService struct {
	repo repository.TimeSlotRepository
}

func NewTimeSlotService(repo repository.TimeSlotRepository) *TimeSlotService {
	return &TimeSlotService{repo: repo}
}

func (s *TimeSlotService) CreateTimeSlot(timeSlot *models.TimeSlot) error {
	if timeSlot.StartTime.After(timeSlot.EndTime) || timeSlot.StartTime.Equal(timeSlot.EndTime) {
		return errors.New("start time must be before end time")
	}
	return s.repo.Create(timeSlot)
}

func (s *TimeSlotService) GetTimeSlot(id uint) (*models.TimeSlot, error) {
	return s.repo.FindByID(id)
}

func (s *TimeSlotService) GetTimeSlotsByEvent(eventID uint) ([]models.TimeSlot, error) {
	return s.repo.FindByEventID(eventID)
}

func (s *TimeSlotService) UpdateTimeSlot(id uint, timeSlot *models.TimeSlot) error {
	if timeSlot.StartTime.After(timeSlot.EndTime) || timeSlot.StartTime.Equal(timeSlot.EndTime) {
		return errors.New("start time must be before end time")
	}
	return s.repo.Update(id, timeSlot)
}

func (s *TimeSlotService) DeleteTimeSlot(id uint) error {
	return s.repo.Delete(id)
}

// UserService handles business logic for users
type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.Create(user)
}

func (s *UserService) GetUser(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) UpdateUser(id uint, user *models.User) error {
	return s.repo.Update(id, user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// AvailabilityService handles business logic for user availability
type AvailabilityService struct {
	repo repository.UserAvailabilityRepository
}

func NewAvailabilityService(repo repository.UserAvailabilityRepository) *AvailabilityService {
	return &AvailabilityService{repo: repo}
}

func (s *AvailabilityService) CreateAvailability(availability *models.UserAvailability) error {
	if availability.StartTime.After(availability.EndTime) || availability.StartTime.Equal(availability.EndTime) {
		return errors.New("start time must be before end time")
	}
	return s.repo.Create(availability)
}

func (s *AvailabilityService) GetUserAvailability(userID, eventID uint) ([]models.UserAvailability, error) {
	return s.repo.FindByUserAndEvent(userID, eventID)
}

func (s *AvailabilityService) UpdateAvailability(id uint, availability *models.UserAvailability) error {
	if availability.StartTime.After(availability.EndTime) || availability.StartTime.Equal(availability.EndTime) {
		return errors.New("start time must be before end time")
	}
	return s.repo.Update(id, availability)
}

func (s *AvailabilityService) DeleteAvailability(id uint) error {
	return s.repo.Delete(id)
}

// RecommendationService handles business logic for generating time slot recommendations
type RecommendationService struct {
	eventRepo        repository.EventRepository
	timeSlotRepo     repository.TimeSlotRepository
	availabilityRepo repository.UserAvailabilityRepository
}

func NewRecommendationService(
	eventRepo repository.EventRepository,
	timeSlotRepo repository.TimeSlotRepository,
	availabilityRepo repository.UserAvailabilityRepository,
) *RecommendationService {
	return &RecommendationService{
		eventRepo:        eventRepo,
		timeSlotRepo:     timeSlotRepo,
		availabilityRepo: availabilityRepo,
	}
}

func (s *RecommendationService) GetRecommendations(eventID uint) ([]models.TimeSlotRecommendation, error) {
	// Get the event to retrieve duration
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, err
	}

	durationMinutes := event.DurationMinutes

	// Get all time slots for the event
	timeSlots, err := s.timeSlotRepo.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}

	// Fetch all availabilities for this event in one query (bulk fetch)
	allAvailabilities, err := s.availabilityRepo.FindByEvent(eventID)
	if err != nil {
		return nil, err
	}

	// Get all users who have provided availability for this event
	users, err := s.availabilityRepo.FindAllUsersByEvent(eventID)
	if err != nil {
		return nil, err
	}

	// Build a map of userID -> slice of availabilities for quick lookup
	availabilityMap := make(map[uint][]models.UserAvailability)
	for _, avail := range allAvailabilities {
		availabilityMap[avail.UserID] = append(availabilityMap[avail.UserID], avail)
	}

	var recommendations []models.TimeSlotRecommendation

	// For each time slot, calculate which users can attend
	for _, slot := range timeSlots {
		var bestMatchingUsers []models.User
		var bestNonMatchingUsers []models.User
		var startOptions []time.Time

		// Check if the slot duration is sufficient for the meeting
		slotDuration := slot.EndTime.Sub(slot.StartTime).Minutes()
		if slotDuration < float64(durationMinutes) {
			continue // Skip this slot if it's too short
		}

		// Calculate the maximum start time within the slot
		maxStartTime := slot.EndTime.Add(-time.Duration(durationMinutes) * time.Minute)

		// Iterate through possible start times at 15-minute intervals
		for startTime := slot.StartTime; !startTime.After(maxStartTime); startTime = startTime.Add(15 * time.Minute) {
			endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)
			var matchingUsers []models.User
			var nonMatchingUsers []models.User

			// Check each user's availabilities from the pre-fetched map
			for _, user := range users {
				availabilities := availabilityMap[user.ID]
				available := false
				for _, avail := range availabilities {
					if !startTime.Before(avail.StartTime) && !endTime.After(avail.EndTime) {
						available = true
						break
					}
				}
				if available {
					matchingUsers = append(matchingUsers, user)
				} else {
					nonMatchingUsers = append(nonMatchingUsers, user)
				}
			}

			// Update best option if current matching count is better
			if len(matchingUsers) > len(bestMatchingUsers) {
				bestMatchingUsers = matchingUsers
				bestNonMatchingUsers = nonMatchingUsers
				startOptions = []time.Time{startTime}
			} else if len(matchingUsers) == len(bestMatchingUsers) && len(matchingUsers) > 0 {
				// If equally good, record additional start option
				startOptions = append(startOptions, startTime)
			}
		}

		// Skip slot if no valid start time is found
		if len(bestMatchingUsers) == 0 {
			continue
		}

		// Calculate matching percentage for this slot
		matchingPercentage := float64(len(bestMatchingUsers)) / float64(len(users)) * 100

		// Append the recommendation for this time slot
		recommendations = append(recommendations, models.TimeSlotRecommendation{
			TimeSlot:           slot,
			MatchingUsers:      bestMatchingUsers,
			NonMatchingUsers:   bestNonMatchingUsers,
			MatchingPercentage: matchingPercentage,
			EventDuration:      durationMinutes,
			StartOptions:       startOptions,
		})
	}

	// Sort recommendations by matching percentage (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].MatchingPercentage > recommendations[j].MatchingPercentage
	})

	return recommendations, nil
}
