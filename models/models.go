package models

import (
	"time"

	"gorm.io/gorm"
)

// Event represents a meeting or event
type Event struct {
	gorm.Model
	Title           string     `json:"title" binding:"required"`
	Description     string     `json:"description"`
	OrganizerId     uint       `json:"organizer_id" binding:"required"`
	DurationMinutes int        `json:"duration_minutes" binding:"required,min=1"`
	TimeSlots       []TimeSlot `json:"time_slots,omitempty" gorm:"foreignKey:EventID"`
}

// TimeSlot represents a potential time for an event
type TimeSlot struct {
	gorm.Model
	EventID   uint      `json:"event_id" gorm:"index"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

// User represents a user of the system
type User struct {
	gorm.Model
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email" gorm:"uniqueIndex"`
	Timezone string `json:"timezone" binding:"required"`
}

// UserAvailability represents a user's availability for an event
type UserAvailability struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"index"`
	EventID   uint      `json:"event_id" gorm:"index"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

// TimeSlotRecommendation represents a recommended time slot with participant info
type TimeSlotRecommendation struct {
	TimeSlot           TimeSlot    `json:"time_slot"`
	MatchingUsers      []User      `json:"matching_users"`
	NonMatchingUsers   []User      `json:"non_matching_users"`
	MatchingPercentage float64     `json:"matching_percentage"`
	EventDuration      int         `json:"event_duration"`
	StartOptions       []time.Time `json:"start_options,omitempty"`
}
