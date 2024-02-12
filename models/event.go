package models

import (
	"fmt"

	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EventRepo wraps the db connection pool in a custom type
// This approach fits nicely to perform unit tests since we can reference EventRepo in the application code
// with an interface
type EventRepo struct {
	DB *gorm.DB
}

const (
	ExternalDomain string = "EXTERNAL"
	InternalDomain string = "ORANMGR"
)

// EventSeverity type reflects represent event classification
type EventSeverity int

const (
	EventSeverityCleared EventSeverity = iota + 1
	EventSeverityIndeterminate
	EventSeverityWarning
	EventSeverityMinor
	EventSeverityMajor
	EventSeverityCritical
	eventSeverityUnknown
)

func (eS EventSeverity) String() string {
	return [...]string{"cleared", "indeterminate", "warning", "minor", "major", "critical"}[eS-1]
}

// Event resemble the DB events table schema
type Event struct {
	gorm.Model
	Username    string
	Activated   time.Time
	Description string
	Modified    time.Time
	AuthnDomain string
	Severity    EventSeverity
}

// TableName returns the Event table name
func (e *Event) TableName() string {
	return "events"
}

// ToString provides a string representation of the Event information
func (e *Event) ToString() string {
	return fmt.Sprintf("id: %d\nusername: %s\nactivated: %s\ndescription: %s\nmodified: %s\nseverity: %s", e.ID, e.Username, e.Activated, e.Description, e.Modified, e.Severity)
}

// Create creates a new event into the DB
// the function returns error if the event with a specific ID is already present within the DB
func (eR *EventRepo) Create(e *Event) error {
	if e.ID != 0 {
		res := eR.DB.First(&e)
		if res.RowsAffected != 0 {
			return &UserError{fmt.Sprintf("event ID %d already present in database", e.ID)}
		}
	}

	res := eR.DB.Create(&e)
	if res.Error != nil {
		return &DBError{res.Error.Error()}
	}
	return nil
}

func (eR *EventRepo) CreateSuccessfulLoginEvent(username, domain, ip string) error {
	e := &Event{
		Username:    username,
		AuthnDomain: domain,
		Activated:   time.Now(),
		Description: fmt.Sprintf("Login successful from IP %s", ip),
		Modified:    time.Now(),
		Severity:    EventSeverityCleared,
	}
	return eR.Create(e)
}

func (eR *EventRepo) CreateUnsuccessfulLoginEvent(username, domain, ip string) error {
	e := &Event{
		Username:    username,
		AuthnDomain: domain,
		Activated:   time.Now(),
		Description: fmt.Sprintf("Login unsuccessful from IP %s", ip),
		Modified:    time.Now(),
		Severity:    EventSeverityWarning,
	}
	return eR.Create(e)
}

// CreateUserEvent creates a new user event into the DB
func (eR *EventRepo) CreateUserEvent(method, username, domain string) error {
	var description string

	switch method {
	case http.MethodPost:
		description = "Added user: " + username
	case http.MethodPatch:
		description = "Updated user: " + username
	case http.MethodDelete:
		description = "Deleted user: " + username
	default:
		description = "Unknown user operation: " + username
	}

	e := &Event{
		Username:    username,
		Activated:   time.Now(),
		Description: description,
		Modified:    time.Now(),
		AuthnDomain: domain,
		Severity:    EventSeverityCleared,
	}
	return eR.Create(e)
}

// CreateJWTEvent creates a new jwt token creation failure event
func (eR *EventRepo) CreateJWTEvent(username, domain string) error {
	e := &Event{
		Username:    username,
		Activated:   time.Now(),
		Description: "Failed creating new session JWT (JavaScript Object Notation Web Token)",
		Modified:    time.Now(),
		AuthnDomain: domain,
		Severity:    EventSeverityWarning,
	}

	return eR.Create(e)
}

// GetEvents returns the list of events present in DB
func (eR *EventRepo) GetEvents(pageNumber int, pageSize int) []*Event {
	offset := pageSize * (pageNumber - 1)
	var events []*Event
	err := eR.DB.Model(&Event{}).Offset(offset).Limit(pageSize).Order("id desc").Find(&events)
	if err.Error != nil {
		log.Fatalf("error %v", err)
	}

	return events
}

// GetEventsCount returns the number of events for the specified severity
func (eR *EventRepo) GetEventsCount(s EventSeverity) int {
	eventsCount := int64(0)
	eR.DB.Model(&Event{}).Where("severity = ?", s).Count(&eventsCount)
	return int(eventsCount)
}

// SetupAutomaticDeletion uses the gocron package to create a cronjob in order to delete excess events on DB
// it is possible to use both cron syntax or time.Duration ("5m", "10h", ...)
func SetupAutomaticDeletion(db *gorm.DB, schedule string, location *time.Location, maxEventsNumber int64) error {
	s := gocron.NewScheduler(location)

	_, err := cron.ParseStandard(schedule)
	if err != nil {
		_, err = time.ParseDuration(schedule)
		if err != nil {
			return err
		}
		_, err = s.Every(schedule).Do(func() {
			err := deleteOldestEvents(db, maxEventsNumber)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Errorf("failed to delete excess events")
			}
		})
		if err != nil {
			return err
		}
	} else {
		_, err = s.Cron(schedule).Do(func() {
			err := deleteOldestEvents(db, maxEventsNumber)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Errorf("failed to delete excess events")
			}
		})
		if err != nil {
			return err
		}
	}
	s.StartAsync()
	return nil
}

// deleteOldestEvents is the utility function which retrieves and deletes all excess events ids from the database, starting from the oldest.
// if delete_active_events is set to true, all events are retrieved without any checks on their severity
func deleteOldestEvents(database *gorm.DB, maxEventsNumber int64) error {
	var eventsToDelete int
	var eventsNumber int64

	database.Model(&Event{}).Select("id").Count(&eventsNumber)
	if eventsNumber-maxEventsNumber <= 0 {
		return nil
	}
	eventsToDelete = int(eventsNumber - maxEventsNumber)
	var ids []int

	db := database.
		Model(&Event{}).
		Select("id").
		Order("activated").
		Limit(eventsToDelete)

	err := db.Scan(&ids).Error
	if err != nil {
		return err
	}

	if ids == nil {
		return nil
	}

	err = database.Delete(&Event{}, ids).Error
	if err != nil {
		return err
	}
	return nil
}
