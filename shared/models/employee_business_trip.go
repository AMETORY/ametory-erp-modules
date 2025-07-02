package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeBusinessTrip struct {
	shared.BaseModel
	EmployeeID             *string         `json:"employee_id,omitempty" gorm:"type:varchar(255)"`
	Employee               *EmployeeModel  `gorm:"foreignKey:EmployeeID" json:"employee"`
	CompanyID              *string         `json:"company_id,omitempty" gorm:"type:varchar(255)"`
	Company                *CompanyModel   `gorm:"foreignKey:CompanyID" json:"company"`
	Date                   time.Time       `json:"date,omitempty"`
	DepartureDate          *time.Time      `json:"departure_date,omitempty"`
	ArrivalDate            *time.Time      `json:"arrival_date,omitempty"`
	Origin                 string          `json:"origin,omitempty" gorm:"type:varchar(255)"`
	Destination            string          `json:"destination,omitempty" gorm:"type:varchar(255)"`
	OriginCityID           *string         `json:"origin_city_id,omitempty" gorm:"type:varchar(255)"`
	DestinationCityID      *string         `json:"destination_city_id,omitempty" gorm:"type:varchar(255)"`
	BusinessTripNumber     string          `json:"business_trip_number,omitempty" gorm:"type:varchar(255)"`
	BusinessTripPurpose    string          `json:"business_trip_purpose,omitempty" gorm:"type:varchar(255)"`
	Status                 string          `json:"status,omitempty" gorm:"default:'DRAFT';type:varchar(255)"`
	Notes                  string          `json:"notes,omitempty" gorm:"type:text"`
	Remarks                string          `json:"remarks,omitempty" gorm:"type:text"`
	ApproverID             *string         `json:"approver_id,omitempty" gorm:"type:varchar(255)"`
	Approver               *EmployeeModel  `gorm:"foreignKey:ApproverID" json:"approver"`
	IsHotelEnabled         bool            `json:"is_hotel_enabled,omitempty"`
	IsTransportEnabled     bool            `json:"is_transport_enabled,omitempty"`
	TripType               string          `json:"trip_type,omitempty" gorm:"default:'ONE_WAY';type:varchar(255)"`    // 'ONE_WAY', 'ROUND_TRIP'
	TransportType          string          `json:"transport_type,omitempty" gorm:"default:'PLANE';type:varchar(255)"` // 'AIRPLANE', 'TRAIN', 'BUS', etc.
	HotelName              string          `json:"hotel_name,omitempty" gorm:"type:varchar(255)"`
	HotelAddress           string          `json:"hotel_address,omitempty" gorm:"type:varchar(255)"`
	HotelContact           string          `json:"hotel_contact,omitempty" gorm:"type:varchar(255)"`
	HotelPhotoURL          *string         `json:"hotel_photo_url,omitempty" gorm:"type:varchar(255)"`
	HotelLat               *float64        `json:"hotel_lat,omitempty" gorm:"type:DECIMAL(10,8)"`
	HotelLng               *float64        `json:"hotel_lng,omitempty" gorm:"type:DECIMAL(11,8)"`
	TripParticipants       []EmployeeModel `json:"trip_participants" gorm:"many2many:trip_participants;constraint:OnDelete:CASCADE;"`
	HotelBookingFiles      []FileModel     `json:"hotel_booking_files,omitempty" gorm:"-"`
	TransportBookingFiles  []FileModel     `json:"transport_booking_files,omitempty" gorm:"-"`
	DateApprovedOrRejected *time.Time      `json:"date_approved_or_rejected"`
	ApprovalByAdminID      *string         `json:"approval_by_admin_id"`
	ApprovalByAdmin        *UserModel      `json:"approval_by_admin" gorm:"foreignKey:ApprovalByAdminID"`
}

func (e *EmployeeBusinessTrip) BeforeCreate(tx *gorm.DB) error {

	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
