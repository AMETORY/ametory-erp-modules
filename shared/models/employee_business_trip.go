package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type EmployeeBusinessTrip struct {
	shared.BaseModel
	EmployeeID            *string         `json:"employee_id,omitempty" gorm:"type:varchar(255)"`
	Employee              *EmployeeModel  `gorm:"foreignKey:EmployeeID" json:"employee"`
	CompanyID             *string         `json:"company_id,omitempty" gorm:"type:varchar(255)"`
	Company               *CompanyModel   `gorm:"foreignKey:CompanyID" json:"company"`
	DepartureDate         *time.Time      `json:"departure_date,omitempty"`
	ArrivalDate           *time.Time      `json:"arrival_date,omitempty"`
	Origin                string          `json:"origin,omitempty" gorm:"type:varchar(255)"`
	Destination           string          `json:"destination,omitempty" gorm:"type:varchar(255)"`
	OriginCityID          *string         `json:"origin_city_id,omitempty" gorm:"type:varchar(255)"`
	DestinationCityID     *string         `json:"destination_city_id,omitempty" gorm:"type:varchar(255)"`
	BusinessTripNumber    string          `json:"business_trip_number,omitempty" gorm:"type:varchar(255)"`
	BusinessTripPurpose   string          `json:"business_trip_purpose,omitempty" gorm:"type:varchar(255)"`
	StartDate             time.Time       `json:"start_date,omitempty"`
	EndDate               time.Time       `json:"end_date,omitempty"`
	Status                string          `json:"status,omitempty" gorm:"default:'DRAFT';type:varchar(255)"`
	Notes                 string          `json:"notes,omitempty" gorm:"type:text"`
	Remarks               string          `json:"remarks,omitempty" gorm:"type:text"`
	ApproverID            *string         `json:"approver_id,omitempty" gorm:"type:varchar(255)"`
	Approver              *EmployeeModel  `gorm:"foreignKey:ApproverID" json:"approver"`
	IsHotelEnabled        bool            `json:"is_hotel_enabled,omitempty"`
	IsTransportEnabled    bool            `json:"is_transport_enabled,omitempty"`
	TripType              string          `json:"trip_type,omitempty" gorm:"default:'ONE_WAY';type:varchar(255)"`    // 'ONE_WAY', 'ROUND_TRIP'
	TransportType         string          `json:"transport_type,omitempty" gorm:"default:'PLANE';type:varchar(255)"` // 'AIRPLANE', 'TRAIN', 'BUS', etc.
	HotelName             string          `json:"hotel_name,omitempty" gorm:"type:varchar(255)"`
	HotelAddress          string          `json:"hotel_address,omitempty" gorm:"type:varchar(255)"`
	HotelContact          string          `json:"hotel_contact,omitempty" gorm:"type:varchar(255)"`
	TripParticipants      []EmployeeModel `json:"trip_participants" gorm:"many2many:trip_participants;constraint:OnDelete:CASCADE;"`
	HotelBookingFiles     []FileModel     `json:"hotel_booking_files,omitempty" gorm:"-"`
	TransportBookingFiles []FileModel     `json:"transport_booking_files,omitempty" gorm:"-"`
}
