package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DateOnly struct {
	time.Time
}
type TimeOnly struct {
	time.Time
}

func (TimeOnly) GormDataType() string {
	return "time"
}

func (TimeOnly) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "time"
}
func (DateOnly) GormDataType() string {
	return "date"
}

func (DateOnly) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "date"
}

func (timeOnly TimeOnly) Value() (driver.Value, error) {
	if !timeOnly.IsZero() {
		return timeOnly.GetTime().Format("15:04:05"), nil
	} else {
		return nil, nil
	}
}

func (timeOnly *TimeOnly) GetTime() time.Time {
	return timeOnly.Time
}
func (timeOnly *TimeOnly) ToTime(newTime time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	return time.Date(newTime.Year(), newTime.Month(), newTime.Day(), timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), timeOnly.Nanosecond(), loc)
}

func (timeOnly *TimeOnly) Scan(value interface{}) error {
	scanned, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to scan Time value:", value))
	}
	scannedString := string(scanned)
	scannedTime, err := time.Parse("15:04:05", scannedString)
	if err == nil {
		*timeOnly = TimeOnly{scannedTime}
	}
	return err
}

func (date *DateOnly) Scan(value interface{}) error {
	scanned, ok := value.(time.Time)
	if !ok {
		return errors.New(fmt.Sprint("Failed to scan DateOnly value:", value))
	}
	*date = DateOnly{scanned}
	return nil
}

func (timeOnly TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(timeOnly.GetTime().Format("15:04:05"))
}

func (timeOnly *TimeOnly) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	t, err := time.ParseInLocation("15:04:05", s, time.UTC)
	if err != nil {
		return err
	}
	*timeOnly = TimeOnly{t}
	return nil
}
