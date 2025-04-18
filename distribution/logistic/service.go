package logistic

import (
	"errors"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type LogisticService struct {
	db               *gorm.DB
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
}

func NewLogisticService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *LogisticService {
	return &LogisticService{db: db, ctx: ctx, inventoryService: inventoryService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DistributionEventModel{},
		&models.ShipmentModel{},
		&models.ShipmentLegModel{},
		&models.TrackingEventModel{},
		&models.IncidentEventModel{},
		&models.IncidentItem{},
		&models.DistributionEventReport{},
		&models.ShipmentFeedback{},
	)
}

func (s *LogisticService) CreateDistributionEvent(data *models.DistributionEventModel) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}
	return nil
}
func (s *LogisticService) CreateShipment(data *models.ShipmentModel) error {
	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	return nil
}

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

func (s *LogisticService) UpdateIsDelayedForShipment(shipmentID string) error {
	return s.db.Model(&models.ShipmentModel{}).Where("id = ?", shipmentID).
		Where("expected_finish_at < ? AND is_delayed = ? and expected_finish_at is not null", time.Now(), false).
		Update("is_delayed", true).Error
}

func (s *LogisticService) CreateShipmentLeg(data *models.ShipmentLegModel) error {
	if err := s.db.Create(&data).Error; err != nil {
		return err
	}

	return nil
}

func (s *LogisticService) MonitorDistributionEvent(eventID string) (*models.DistributionEventModel, error) {
	distributionEvent := models.DistributionEventModel{}
	if err := s.db.
		Preload("Shipments.ShipmentLegs").
		Preload("Shipments.Items").
		First(&distributionEvent, "id = ?", eventID).Error; err != nil {
		return nil, err
	}

	return &distributionEvent, nil
}

func (s *LogisticService) MonitorShipment(shipmentID string) (*models.ShipmentModel, error) {
	shipment := models.ShipmentModel{}
	if err := s.db.
		Preload("ShipmentLegs").
		Preload("Items").
		First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil, err
	}

	return &shipment, nil
}

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
	if err := s.db.Save(&shipmentLeg).Error; err != nil {
		return err
	}

	if shipmentLeg.ToLocationID == nil {
		return errors.New("to location id is nil")
	}

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

func (s *LogisticService) AddTrackingEvent(shipmentLegID string, status string, latitude float64, longitude float64, notes string) error {
	trackingEvent := models.TrackingEventModel{
		ShipmentLegID: &shipmentLegID,
		Status:        status,
		Latitude:      latitude,
		Longitude:     longitude,
		Timestamp:     time.Now(),
		Notes:         notes,
	}

	if err := s.db.Create(&trackingEvent).Error; err != nil {
		return err
	}

	return nil
}

func (s *LogisticService) CreateShipmentFeedback(data *models.ShipmentFeedback) error {
	return s.db.Create(data).Error
}

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

func (s *LogisticService) GenerateDistributionEventReport(distributionEventID string, reportID *string) (*models.DistributionEventReport, error) {
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
		DistributionEvent:   distributionEvent,
		DistributionEventID: distributionEventID,
	}

	if reportID != nil {
		if err := s.db.First(&report, "id = ?", *reportID).Error; err != nil {
			return nil, err
		}
	}

	report.TotalShipments = (len(distributionEvent.Shipments))
	report.TotalDestinations = 0
	for _, shipment := range distributionEvent.Shipments {
		report.TotalDestinations += len(shipment.ShipmentLegs)
	}
	report.TotalItems = 0
	for _, shipment := range distributionEvent.Shipments {
		report.TotalItems += (len(shipment.Items))
	}
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
		if shipment.Status == "DELIVERED" {
			report.FinishedShipments += 1
		} else if shipment.Status == "IN_DELIVERY" || shipment.Status == "PROCESSING" {
			report.ProcessingShipments += 1
		} else if shipment.Status == "READY_TO_SHIP" {
			report.ReadyToShip += 1
		}

		for _, incidentEvent := range shipment.IncidentEvents {
			if incidentEvent.EventType == "LOST" {
				report.LostItems += (len(incidentEvent.Items))
			} else if incidentEvent.EventType == "DAMAGE" {
				report.DamagedItems += (len(incidentEvent.Items))
			}
		}
		report.FeedbackCount = len(shipment.Feedbacks)
	}

	err := s.db.Save(&report).Error

	return &report, err
}

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
