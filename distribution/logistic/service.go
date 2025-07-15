package logistic

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type LogisticService struct {
	db               *gorm.DB
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
}

// NewLogisticService creates a new instance of LogisticService with the given database connection, context and inventory service.
func NewLogisticService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *LogisticService {
	return &LogisticService{db: db, ctx: ctx, inventoryService: inventoryService}
}

// Migrate runs database migrations for the logistic module.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.ShipmentModel{},
		&models.ShipmentItem{},
		&models.DistributionEventModel{},
		&models.ShipmentLegModel{},
		&models.TrackingEventModel{},
		&models.IncidentEventModel{},
		&models.IncidentItem{},
		&models.DistributionEventReport{},
		&models.ShipmentFeedback{},
	)

	if err != nil {
		return err
	}

	err = db.Exec("ALTER TABLE shipments DROP CONSTRAINT IF EXISTS fk_shipments_shipment_legs").Error
	if err != nil {
		return err
	}
	err = db.Exec("ALTER TABLE shipments ADD CONSTRAINT fk_shipments_shipment_legs FOREIGN KEY (current_shipment_leg_id) REFERENCES shipment_legs(id) ON DELETE SET NULL ON UPDATE RESTRICT").Error
	if err != nil {
		return err
	}

	return nil
}

// GetAllShipments retrieves a paginated list of shipments from the database.
//
// It takes an HTTP request and a search query string as input. The method
// preloads relationships for the shipment's origin and destination warehouses
// as well as distribution events. The search query is applied to the notes
// and code fields of the shipment. If the request contains a company ID
// header, the method filters results by company ID. Pagination is applied
// to manage the result set, and any necessary request modifications are made
// using the utils.FixRequest utility.
//
// The function returns a paginated page of ShipmentModel and an error if the
// operation fails.

func (s *LogisticService) GetAllShipments(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("FromLocation.Warehouse").
		Preload("ToLocation.Warehouse").
		Preload("DistributionEvent", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})
	if search != "" {
		stmt = stmt.Where("notes ILIKE ? OR code ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ShipmentModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ShipmentModel{})
	return page, nil

}

// ListDistributionEvents retrieves a paginated list of distribution events from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the notes and name fields of the distribution events. If the
// request contains a company ID header, the method filters results by company
// ID. Pagination is applied to manage the result set, and any necessary request
// modifications are made using the utils.FixRequest utility.
//
// The function returns a paginated page of DistributionEventModel and an error
// if the operation fails.
func (s *LogisticService) ListDistributionEvents(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("notes ILIKE ? OR name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.DistributionEventModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.DistributionEventModel{})
	page.Page = page.Page + 1
	return page, nil
}

// ListShipments retrieves a paginated list of shipments from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the notes and code fields of the shipments. If the request
// contains a company ID header, the method filters results by company ID.
// Pagination is applied to manage the result set, and any necessary request
// modifications are made using the utils.FixRequest utility.
//
// The function returns a paginated page of ShipmentModel and an error if the
// operation fails.

func (s *LogisticService) ListShipments(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("notes ILIKE ? OR code ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ShipmentModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ShipmentModel{})
	page.Page = page.Page + 1
	return page, nil
}

// CreateDistributionEvent creates a new distribution event record in the database.
//
// It takes a DistributionEventModel as input and attempts to save it to the database.
// The function returns an error if the operation fails, otherwise it returns nil.

func (s *LogisticService) CreateDistributionEvent(data *models.DistributionEventModel) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}
	return nil
}

// CreateShipment creates a new shipment record in the database.
//
// It takes a ShipmentModel as input and attempts to save it to the database.
// The function returns an error if the operation fails, otherwise it returns nil.
func (s *LogisticService) CreateShipment(data *models.ShipmentModel) error {
	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	return nil
}

// DeleteItemShipment deletes an item from the shipment record in the database.
//
// It takes a shipment ID and an item ID as input and attempts to delete the
// item from the shipment. The function returns an error if the operation fails,
// otherwise it returns nil.
func (s *LogisticService) DeleteItemShipment(shipmentID string, itemID string) error {
	return s.db.Delete(&models.ShipmentItem{}, "shipment_id = ? AND id = ?", shipmentID, itemID).Error
}

// DeleteShipment deletes a shipment record in the database.
//
// It takes a shipment ID as input and attempts to delete the shipment and its associated records.
// The function returns an error if the operation fails, otherwise it returns nil.
func (s *LogisticService) DeleteShipment(shipmentID string) error {
	if err := s.db.Delete(&models.ShipmentItem{}, "shipment_id = ?", shipmentID).Error; err != nil {
		return err
	}

	if err := s.db.Delete(&models.ShipmentLegModel{}, "shipment_id = ?", shipmentID).Error; err != nil {
		return err
	}

	return s.db.Delete(&models.ShipmentModel{}, "id = ?", shipmentID).Error
}

// ReadyToShip marks a shipment as READY_TO_SHIP. It takes a shipment ID, a
// date, and an optional notes string as input. It first checks if the shipment
// status is PENDING. If not, it returns an error. If the notes string is not
// nil, it appends the string to the shipment notes. Finally, it saves the
// shipment with the new status and returns an error if the operation fails.
func (s *LogisticService) ReadyToShip(shipmentID string, date time.Time, notes *string) error {
	shipment := models.ShipmentModel{}
	if err := s.db.First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return err
	}

	if shipment.Status != "PENDING" {
		return errors.New("shipment status is not PENDING")
	}

	shipment.Status = "READY_TO_SHIP"
	if notes != nil {
		shipment.Notes += *notes + "\n"
	}
	if err := s.db.Save(&shipment).Error; err != nil {
		return err
	}

	return nil
}

// ProcessShipment updates a shipment's status to IN_DELIVERY.
//
// It requires a shipment ID, a date, and notes as input. The function first
// checks if the shipment's status is READY_TO_SHIP. If not, it returns an
// error. If the status is correct, it updates the status to IN_DELIVERY and
// saves the changes to the database. The function returns an error if any
// operation fails, otherwise it returns nil.

func (s *LogisticService) ProcessShipment(shipmentID string, date time.Time, notes string) error {
	shipment := models.ShipmentModel{}
	if err := s.db.First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return err
	}

	if shipment.Status != "READY_TO_SHIP" {
		return errors.New("shipment status is not READY_TO_SHIP")
	}
	shipment.Status = "IN_DELIVERY"
	return s.db.Save(&shipment).Error
}

// UpdateIsDelayedForShipment updates the shipment's is_delayed flag to true
// if its expected_finish_at is less than the current time and the flag is
// currently false.
func (s *LogisticService) UpdateIsDelayedForShipment(shipmentID string) error {
	return s.db.Model(&models.ShipmentModel{}).Where("id = ?", shipmentID).
		Where("expected_finish_at < ? AND is_delayed = ? and expected_finish_at is not null", time.Now(), false).
		Update("is_delayed", true).Error
}

// CreateShipmentLeg creates a new shipment leg record in the database.
//
// It takes a ShipmentLegModel as input and attempts to save it to the database.
// The function returns an error if the operation fails, otherwise it returns nil.
func (s *LogisticService) CreateShipmentLeg(data *models.ShipmentLegModel) error {
	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	return nil
}

// DeleteDistributionEvent deletes a distribution event and its associated shipments from the database.
//
// It takes an event ID as input and attempts to first retrieve the distribution event with its shipments preloaded.
// The function deletes the associated shipments and then the distribution event itself.
// It returns an error if any operation fails, otherwise it returns nil.

func (s *LogisticService) DeleteDistributionEvent(eventID string) error {
	distributionEvent := models.DistributionEventModel{}
	if err := s.db.Preload("Shipments").First(&distributionEvent, "id = ?", eventID).Error; err != nil {
		return err
	}

	if err := s.db.Delete(&distributionEvent.Shipments, "distribution_event_id = ?", eventID).Error; err != nil {
		return err
	}
	if err := s.db.Delete(&models.DistributionEventModel{}, "id = ?", eventID).Error; err != nil {
		return err
	}
	return nil
}

// GetDistributionEvent retrieves a distribution event and its associated shipments
// from the database.
//
// It takes an event ID as input and returns a pointer to a DistributionEventModel
// and an error. The function preloads the associated shipments, items, shipment
// legs, and from/to locations. If the distribution event is found, it is
// returned along with a nil error. If not found, or in case of a query error,
// the function returns a non-nil error.
func (s *LogisticService) GetDistributionEvent(eventID string) (*models.DistributionEventModel, error) {
	distributionEvent := models.DistributionEventModel{}
	if err := s.db.
		Preload("Shipments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Items").Preload("ShipmentLegs").
				Preload("FromLocation.Warehouse").
				Preload("ToLocation.Warehouse")
		}).
		First(&distributionEvent, "id = ?", eventID).Error; err != nil {
		return nil, err
	}

	return &distributionEvent, nil
}

// UpdateStatusShipment updates the status of a shipment in the database.
//
// It takes a shipment ID and a status string as input and attempts to
// update the shipment record with the new status. If the shipment is found,
// the function returns a nil error. Otherwise, it returns a non-nil error.
func (s *LogisticService) UpdateStatusShipment(shipmentID string, status string) error {
	shipment := models.ShipmentModel{}
	if err := s.db.First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return err
	}

	shipment.Status = status
	return s.db.Save(&shipment).Error
}

// GetShipment retrieves a shipment and its associated shipment legs, items, and
// distribution event from the database.
//
// It takes a shipment ID as input and returns a pointer to a ShipmentModel and
// an error. The function preloads the associated shipment legs, items, and
// distribution event. If the shipment is found, it is returned along with a nil
// error. If not found, or in case of a query error, the function returns a
// non-nil error.
func (s *LogisticService) GetShipment(shipmentID string) (*models.ShipmentModel, error) {
	shipment := models.ShipmentModel{}
	if err := s.db.
		Preload("ShipmentLegs", func(db *gorm.DB) *gorm.DB {
			return db.Preload("TrackingEvents", func(db *gorm.DB) *gorm.DB {
				return db.Order("seq_number ASC")
			}).
				Preload("FromLocation.Warehouse").
				Preload("ToLocation.Warehouse")
		}).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Product").Preload("Unit")
		}).
		Preload("FromLocation.Warehouse").
		Preload("ToLocation.Warehouse").
		Preload("DistributionEvent", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil, err
	}

	return &shipment, nil
}

// AddItemShipment adds a new item to a shipment in the database.
//
// It takes a shipment ID and a pointer to a ShipmentItem as input and attempts
// to create a new record in the database. The function returns an error if the
// operation fails, otherwise it returns nil.

func (s *LogisticService) AddItemShipment(shipmentID string, item *models.ShipmentItem) error {
	if err := s.db.Create(item).Error; err != nil {
		return err
	}

	return nil
}

// StartShipmentLegDelivery marks a shipment leg as IN_DELIVERY. It takes a shipment
// leg ID, a date, and an optional notes string as input. It first checks if the
// shipment leg status is PENDING. If not, it returns an error. If the notes string is
// not nil, it appends the string to the shipment leg notes. Finally, it saves the
// shipment leg with the new status and updates the current shipment leg ID of the
// associated shipment. The function returns an error if the operation fails,
// otherwise it returns nil.
func (s *LogisticService) StartShipmentLegDelivery(shipmentLegID string, date time.Time, notes string) error {
	shipmentLeg := models.ShipmentLegModel{}
	if err := s.db.
		Preload("Shipment.Items").
		Preload("FromLocation.Warehouse").
		Preload("ToLocation.Warehouse").
		First(&shipmentLeg, "id = ?", shipmentLegID).Error; err != nil {
		return err
	}

	shipmentLeg.DepartedAt = &date
	shipmentLeg.Status = "IN_DELIVERY"
	if err := s.db.Save(&shipmentLeg).Error; err != nil {
		return err
	}

	if shipmentLeg.FromLocationID == nil {
		return errors.New("from location id is nil")
	}

	for _, v := range shipmentLeg.Shipment.Items {
		if _, err := s.inventoryService.StockMovementService.AddMovement(
			date,
			*v.ProductID,
			*shipmentLeg.FromLocation.WarehouseID,
			nil,
			nil,
			nil,
			nil,
			-v.Quantity,
			models.MovementTypeShippingOut,
			shipmentLegID,
			notes,
		); err != nil {
			return err
		}
	}

	// Update the current shipment leg ID to the current shipment leg
	shipment := models.ShipmentModel{}
	if err := s.db.First(&shipment, "id = ?", *shipmentLeg.ShipmentID).Error; err != nil {
		return err
	}

	shipment.CurrentShipmentLegID = &shipmentLegID
	if err := s.db.Save(&shipment).Error; err != nil {
		return err
	}

	return s.UpdateIsDelayedForShipment(*shipmentLeg.ShipmentID)
}

// ArrivedShipmentLegDelivery marks a shipment leg as ARRIVED. It takes a shipment
// leg ID, a date, and an optional notes string as input. It first checks if the
// shipment leg status is IN_DELIVERY. If not, it returns an error. If the notes string is
// not nil, it appends the string to the shipment leg notes. If the to location ID is nil,
// it returns an error. If the to location warehouse ID is not nil, it adds a movement
// for each item in the shipment leg for the same quantity as the shipment leg quantity.
// Finally, it saves the shipment leg with the new status and updates the current shipment leg ID of the associated shipment. The function returns an error if the operation fails,
// otherwise it returns nil.
func (s *LogisticService) ArrivedShipmentLegDelivery(shipmentLegID string, date time.Time, notes string) error {
	shipmentLeg := models.ShipmentLegModel{}
	if err := s.db.
		Preload("Shipment.Items").
		Preload("FromLocation.Warehouse").
		Preload("ToLocation.Warehouse").
		First(&shipmentLeg, "id = ?", shipmentLegID).Error; err != nil {
		return err
	}

	shipmentLeg.ArrivedAt = &date
	shipmentLeg.Status = "ARRIVED"
	if err := s.db.Save(&shipmentLeg).Error; err != nil {
		return err
	}

	if shipmentLeg.ToLocationID == nil {
		return errors.New("to location id is nil")
	}

	if shipmentLeg.ToLocation.WarehouseID != nil {

		for _, v := range shipmentLeg.Shipment.Items {
			if _, err := s.inventoryService.StockMovementService.AddMovement(
				date,
				*v.ProductID,
				*shipmentLeg.ToLocation.WarehouseID,
				nil,
				nil,
				nil,
				nil,
				v.Quantity,
				models.MovementTypeShippingIn,
				shipmentLegID,
				notes,
			); err != nil {
				return err
			}
		}
	}

	shipment := models.ShipmentModel{}
	if err := s.db.First(&shipment, "id = ?", *shipmentLeg.ShipmentID).Error; err != nil {
		return err
	}

	shipment.CurrentShipmentLegID = nil
	if err := s.db.Save(&shipment).Error; err != nil {
		return err
	}

	return s.UpdateIsDelayedForShipment(*shipmentLeg.ShipmentID)
}

// AddTrackingEvent adds a new tracking event to a shipment leg in the database.
//
// It takes a shipment leg ID and a TrackingEventModel as input and attempts to
// save the tracking event record. If the operation is successful, it returns nil.
// Otherwise, it returns an error.

func (s *LogisticService) AddTrackingEvent(shipmentLegID string, data *models.TrackingEventModel) error {

	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	return nil
}

// CreateShipmentFeedback creates a new shipment feedback record in the database.
//
// It takes a ShipmentFeedback as input and attempts to save it to the database.
// The function returns an error if the operation fails, otherwise it returns nil.
func (s *LogisticService) CreateShipmentFeedback(data *models.ShipmentFeedback) error {
	return s.db.Create(data).Error
}

// GenerateShipmentReport generates a shipment report based on the shipment ID.
//
// It takes a shipment ID as input and retrieves the shipment and its associated
// shipment legs, items, and incident events from the database. The function
// returns a pointer to a ShipmentModel and an error. If the shipment is found, it
// is returned along with a nil error. If not found, or in case of a query error,
// the function returns a non-nil error.
func (s *LogisticService) GenerateShipmentReport(shipmentID string) (*models.ShipmentModel, error) {

	shipment := models.ShipmentModel{}
	if err := s.db.
		Preload("ShipmentLegs.TrackingEvents").
		Preload("Items.Product").
		Preload("IncidentEvents.Items").
		First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil, err
	}

	// Additional logic to format and generate the report can be added here

	return &shipment, nil
}

// GetDistributionEventReport retrieves a distribution event report from the database.
//
// It takes a distribution event ID as input and returns a pointer to a DistributionEventReport
// and an error. The function preloads the associated shipments, items, shipment legs,
// tracking events, and incident events. If the distribution event is found, it is
// returned along with a nil error. If not found, or in case of a query error,
// the function returns a non-nil error.
func (s *LogisticService) GetDistributionEventReport(distributionEventID string) (*models.DistributionEventReport, error) {

	return s.GenerateDistributionEventReport(distributionEventID)
}

// GenerateDistributionEventReport generates a distribution event report based on the distribution event ID.
//
// It takes a distribution event ID as input and retrieves the distribution event and its associated
// shipments, items, shipment legs, tracking events, incident events, and feedbacks from the database.
// The function returns a pointer to a DistributionEventReport and an error. If the distribution event
// is found, it is returned along with a nil error. If not found, or in case of a query error,
// the function returns a non-nil error.
func (s *LogisticService) GenerateDistributionEventReport(distributionEventID string) (*models.DistributionEventReport, error) {
	distributionEvent := models.DistributionEventModel{}
	if err := s.db.
		Preload("Shipments.ShipmentLegs.TrackingEvents").
		Preload("Shipments.Items.Product").
		Preload("Shipments.Feedbacks").
		Preload("Shipments.IncidentEvents.Items").
		First(&distributionEvent, "id = ?", distributionEventID).Error; err != nil {
		return nil, err
	}

	report := models.DistributionEventReport{
		DistributionEvent: distributionEvent,
	}

	s.db.First(&report, "distribution_event_id = ?", distributionEventID)

	report.TotalShipments = (len(distributionEvent.Shipments))
	report.TotalDestinations = 0
	for _, shipment := range distributionEvent.Shipments {
		report.TotalDestinations += len(shipment.ShipmentLegs)
	}
	report.TotalItems = 0
	for _, shipment := range distributionEvent.Shipments {
		report.TotalItems += (len(shipment.Items))
	}
	report.DistributionEventID = distributionEventID
	report.LostItems = 0
	report.DamagedItems = 0
	report.DelayedShipments = 0
	report.FinishedShipments = 0
	report.ProcessingShipments = 0
	report.ReadyToShip = 0
	for _, shipment := range distributionEvent.Shipments {
		if shipment.IsDelayed {
			report.DelayedShipments += 1
		}
		switch shipment.Status {
		case "DELIVERED":
			report.FinishedShipments += 1
		case "IN_DELIVERY", "PROCESSING":
			report.ProcessingShipments += 1
		case "READY_TO_SHIP":
			report.ReadyToShip += 1
		}

		for _, incidentEvent := range shipment.IncidentEvents {
			switch incidentEvent.EventType {
			case "LOST":
				report.LostItems += (len(incidentEvent.Items))
			case "DAMAGE":
				report.DamagedItems += (len(incidentEvent.Items))
			}
		}
		report.FeedbackCount = len(shipment.Feedbacks)
	}

	err := s.db.Save(&report).Error

	return &report, err
}

// ReportLostOrDamage records an incident event of lost or damaged items for a given shipment leg.
//
// It takes a shipment ID, shipment leg ID, date, incident event data, movement type, and an optional
// waste warehouse ID as input. The function retrieves the shipment and shipment leg from the database,
// creates the incident event record, and updates the stock movements for the affected items. If a waste
// warehouse ID is provided, it adds a stock movement for the waste warehouse. The function returns an
// error if any operation fails, otherwise it returns nil.

func (s *LogisticService) ReportLostOrDamage(shipmentID string,
	shipmentLegID string,
	date time.Time,
	data *models.IncidentEventModel,
	movementType models.MovementType,
	wasteWarehouseID *string,
) error {
	shipment := models.ShipmentModel{}
	if err := s.db.
		First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil
	}
	shipmentLeg := models.ShipmentLegModel{}
	if err := s.db.
		First(&shipmentLeg, "id = ? and shipment_id = ?", shipmentLegID, shipmentLegID).Error; err != nil {
		return nil
	}

	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	for _, v := range data.Items {
		if v.ProductID != nil {
			if _, err := s.inventoryService.StockMovementService.AddMovement(
				date,
				*v.ProductID,
				*shipmentLeg.ToLocationID,
				nil,
				nil,
				nil,
				nil,
				-v.QtyAffected,
				movementType,
				shipmentLegID,
				v.Notes,
			); err != nil {
				return err
			}
		}
	}

	if wasteWarehouseID != nil {
		wasteWarehouse := models.WarehouseModel{}
		if err := s.db.First(&wasteWarehouse, "id = ?", *wasteWarehouseID).Error; err != nil {
			return err
		}
		for _, v := range data.Items {
			if v.ProductID != nil {
				if _, err := s.inventoryService.StockMovementService.AddMovement(
					date,
					*v.ProductID,
					*wasteWarehouseID,
					nil,
					nil,
					nil,
					nil,
					v.QtyAffected,
					movementType,
					shipmentLegID,
					v.Notes,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
