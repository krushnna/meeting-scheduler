package repository

import (
	"github.com/krushnna/meeting-scheduler/models"
	"gorm.io/gorm"
)

// EventRepository interface defines methods for Event operations
type EventRepository interface {
	Create(event *models.Event) error
	FindByID(id uint) (*models.Event, error)
	FindAll() ([]models.Event, error)
	FindAllWithPagination(limit, offset int) ([]models.Event, error)
	Update(id uint, event *models.Event) error
	Delete(id uint) error
}

// EventRepositoryImpl implements EventRepository
type EventRepositoryImpl struct {
	db *gorm.DB
}

func (r *EventRepositoryImpl) FindAllWithPagination(limit, offset int) ([]models.Event, error) {
	var events []models.Event
	result := r.db.Limit(limit).Offset(offset).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &EventRepositoryImpl{db: db}
}

func (r *EventRepositoryImpl) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *EventRepositoryImpl) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	result := r.db.First(&event, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &event, nil
}

func (r *EventRepositoryImpl) FindAll() ([]models.Event, error) {
	var events []models.Event
	result := r.db.Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (r *EventRepositoryImpl) Update(id uint, event *models.Event) error {
	return r.db.Model(&models.Event{}).Where("id = ?", id).Updates(event).Error
}

func (r *EventRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.Event{}, id).Error
}

// TimeSlotRepository interface defines methods for TimeSlot operations
type TimeSlotRepository interface {
	Create(timeSlot *models.TimeSlot) error
	FindByID(id uint) (*models.TimeSlot, error)
	FindByEventID(eventID uint) ([]models.TimeSlot, error)
	Update(id uint, timeSlot *models.TimeSlot) error
	Delete(id uint) error
}

// TimeSlotRepositoryImpl implements TimeSlotRepository
type TimeSlotRepositoryImpl struct {
	db *gorm.DB
}

func NewTimeSlotRepository(db *gorm.DB) TimeSlotRepository {
	return &TimeSlotRepositoryImpl{db: db}
}

func (r *TimeSlotRepositoryImpl) Create(timeSlot *models.TimeSlot) error {
	return r.db.Create(timeSlot).Error
}

func (r *TimeSlotRepositoryImpl) FindByID(id uint) (*models.TimeSlot, error) {
	var timeSlot models.TimeSlot
	result := r.db.First(&timeSlot, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &timeSlot, nil
}

func (r *TimeSlotRepositoryImpl) FindByEventID(eventID uint) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	result := r.db.Where("event_id = ?", eventID).Find(&timeSlots)
	if result.Error != nil {
		return nil, result.Error
	}
	return timeSlots, nil
}

func (r *TimeSlotRepositoryImpl) Update(id uint, timeSlot *models.TimeSlot) error {
	return r.db.Model(&models.TimeSlot{}).Where("id = ?", id).Updates(timeSlot).Error
}

func (r *TimeSlotRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.TimeSlot{}, id).Error
}

// UserRepository interface defines methods for User operations
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindAll() ([]models.User, error)
	Update(id uint, user *models.User) error
	Delete(id uint) error
}

// UserRepositoryImpl implements UserRepository
type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepositoryImpl) FindByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindAll() ([]models.User, error) {
	var users []models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepositoryImpl) Update(id uint, user *models.User) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

func (r *UserRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// UserAvailabilityRepository interface defines methods for UserAvailability operations
type UserAvailabilityRepository interface {
	Create(availability *models.UserAvailability) error
	FindByID(id uint) (*models.UserAvailability, error)
	FindByUserAndEvent(userID, eventID uint) ([]models.UserAvailability, error)
	FindAllUsersByEvent(eventID uint) ([]models.User, error)
	Update(id uint, availability *models.UserAvailability) error
	Delete(id uint) error
	// New method: fetch all availabilities for an event in one query
	FindByEvent(eventID uint) ([]models.UserAvailability, error)
}

// UserAvailabilityRepositoryImpl implements UserAvailabilityRepository
type UserAvailabilityRepositoryImpl struct {
	db *gorm.DB
}

func NewUserAvailabilityRepository(db *gorm.DB) UserAvailabilityRepository {
	return &UserAvailabilityRepositoryImpl{db: db}
}

func (r *UserAvailabilityRepositoryImpl) Create(availability *models.UserAvailability) error {
	return r.db.Create(availability).Error
}

func (r *UserAvailabilityRepositoryImpl) FindByID(id uint) (*models.UserAvailability, error) {
	var availability models.UserAvailability
	result := r.db.First(&availability, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &availability, nil
}

func (r *UserAvailabilityRepositoryImpl) FindByUserAndEvent(userID, eventID uint) ([]models.UserAvailability, error) {
	var availabilities []models.UserAvailability
	result := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).Find(&availabilities)
	if result.Error != nil {
		return nil, result.Error
	}
	return availabilities, nil
}

func (r *UserAvailabilityRepositoryImpl) FindAllUsersByEvent(eventID uint) ([]models.User, error) {
	var users []models.User
	result := r.db.
		Joins("JOIN user_availabilities ON users.id = user_availabilities.user_id").
		Where("user_availabilities.event_id = ?", eventID).
		Group("users.id").
		Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserAvailabilityRepositoryImpl) Update(id uint, availability *models.UserAvailability) error {
	return r.db.Model(&models.UserAvailability{}).Where("id = ?", id).Updates(availability).Error
}

func (r *UserAvailabilityRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.UserAvailability{}, id).Error
}

func (r *UserAvailabilityRepositoryImpl) FindByEvent(eventID uint) ([]models.UserAvailability, error) {
	var availabilities []models.UserAvailability
	result := r.db.Where("event_id = ?", eventID).Find(&availabilities)
	if result.Error != nil {
		return nil, result.Error
	}
	return availabilities, nil
}
