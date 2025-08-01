package contact

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ContactService struct {
	ctx            *context.ERPContext
	CompanyService *company.CompanyService
}

func NewContactService(ctx *context.ERPContext, companyService *company.CompanyService) *ContactService {
	var contactService = ContactService{ctx: ctx, CompanyService: companyService}
	if !ctx.SkipMigration {
		if err := contactService.Migrate(); err != nil {
			panic(err)
		}
	}
	return &contactService
}

// CreateContact creates a new contact.
//
// It checks if a contact with the same phone number already exists in the same company.
// If it does, it returns an error.
//
// If the contact does not exist, it creates a new contact with the given data.
func (s *ContactService) CreateContact(data *models.ContactModel) error {
	if data.Phone != nil {
		var existingContact models.ContactModel
		if err := s.ctx.DB.Where("phone = ? and company_id = ?", data.Phone, *data.CompanyID).First(&existingContact).Error; err == nil {
			return errors.New("contact with this phone number already exists")
		}
	}
	if data.Email != "" {
		var existingContact models.ContactModel
		if err := s.ctx.DB.Where("email = ? and company_id = ?", data.Email, *data.CompanyID).First(&existingContact).Error; err == nil {
			return errors.New("contact with this email already exists")
		}
	}
	// Check if a contact with the same phone number already exists

	return s.ctx.DB.Create(data).Error
}

// CreateContactFromUser creates a new contact from a user if no contact with the same email exists.
//
// If the contact does not exist, it creates a new contact with the given data.
// If the contact does exist, it returns the existing contact.
func (s *ContactService) CreateContactFromUser(user *models.UserModel, code string, isCustomer, isVendor, isSupplier bool, companyID *string) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.Where("user_id = ?", user.ID).First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			contact = models.ContactModel{
				Code:       code,
				Name:       user.FullName,
				Phone:      user.PhoneNumber,
				Address:    user.Address,
				Email:      user.Email,
				IsCustomer: isCustomer,
				IsVendor:   isVendor,
				IsSupplier: isSupplier,
				UserID:     &user.ID,
				CompanyID:  companyID,
			}
			if err := s.ctx.DB.Create(&contact).Error; err != nil {
				return nil, err
			}
		}
	}

	return &contact, nil
}

// GetContactByPhone retrieves a contact by phone number.
//
// It returns an error if the contact is not found.
func (s *ContactService) GetContactByPhone(phone string, companyID string) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.Where("phone = ? and company_id = ?", phone, companyID).First(&contact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}
	return &contact, nil
}

// GetContactByID retrieves a contact by ID.
//
// It returns an error if the contact is not found.
func (s *ContactService) GetContactByID(id string) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.First(&contact, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}
	if contact.DebtLimit > 0 {
		contact.DebtLimitRemain = contact.DebtLimit - s.GetTotalDebt(&contact)
	}
	if contact.ReceivablesLimit > 0 {
		contact.ReceivablesLimitRemain = contact.ReceivablesLimit - s.GetTotalReceivable(&contact)
	}

	return &contact, nil
}

// UpdateContactRoles updates the roles (is_customer, is_vendor, is_supplier) of a contact.
func (s *ContactService) UpdateContactRoles(id string, isCustomer, isVendor, isSupplier bool) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.First(&contact, id).Error; err != nil {
		return nil, err
	}

	contact.IsCustomer = isCustomer
	contact.IsVendor = isVendor
	contact.IsSupplier = isSupplier

	if err := s.ctx.DB.Save(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetContactsByRole retrieves all contacts based on the specified roles (customer, vendor, supplier).
func (s *ContactService) GetContactsByRole(isCustomer, isVendor, isSupplier bool) ([]models.ContactModel, error) {
	var contacts []models.ContactModel
	query := s.ctx.DB.Where("is_customer = ? AND is_vendor = ? AND is_supplier = ?", isCustomer, isVendor, isSupplier)
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// DeleteContact removes a contact based on the given ID.
func (s *ContactService) DeleteContact(id string) error {
	if err := s.ctx.DB.Delete(&models.ContactModel{}, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contact not found")
		}
		return err
	}
	return nil
}

// UpdateContact updates the information of a contact.
func (s *ContactService) UpdateContact(id string, data *models.ContactModel) (*models.ContactModel, error) {
	if err := s.ctx.DB.Where("id = ?", id).Updates(&data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

// GetContacts mengambil semua contact dengan pagination
func (s *ContactService) GetContacts(request http.Request, search string, isCustomer, isVendor, isSupplier *bool) (paginate.Page, error) {

	pg := paginate.New()
	stmt := s.ctx.DB.Select("contacts.id, contacts.name, contacts.email, contacts.phone, contacts.connection_type, contacts.custom_data").Preload("Tags").Preload("Products")
	if search != "" || request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.
			Joins("LEFT JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
			Joins("LEFT JOIN tags ON tags.id = contact_tags.tag_model_id")
	}
	if search != "" {

		stmt = stmt.Where("contacts.name ILIKE ? OR contacts.email ILIKE ? OR contacts.phone ILIKE ? OR contacts.address ILIKE ? OR tags.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)

	}

	if isCustomer != nil || isVendor != nil || isSupplier != nil {
		sbWhere := s.ctx.DB
		if isCustomer != nil {
			sbWhere = sbWhere.Or("is_customer = ?", *isCustomer)
		}
		if isVendor != nil {
			sbWhere = sbWhere.Or("is_vendor = ?", *isVendor)
		}
		if isSupplier != nil {
			sbWhere = sbWhere.Or("is_supplier = ?", *isSupplier)
		}
		stmt = stmt.Where(sbWhere)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("contacts.company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	}
	if request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.Where("tags.id in (?)", strings.Split(request.URL.Query().Get("tag_ids"), ","))
	}

	stmt = stmt.Group("contacts.id, contacts.name, contacts.email, contacts.phone")
	stmt = stmt.Model(&models.ContactModel{})
	fmt.Println("REQUEST CONTACT", request)
	utils.LogJson(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ContactModel{})
	utils.FixRequest(&request)
	page.Page = page.Page + 1

	items := page.Items.(*[]models.ContactModel)
	newItems := make([]models.ContactModel, 0)
	for _, item := range *items {
		if item.DebtLimit > 0 {
			var total = s.GetTotalDebt(&item)
			item.TotalDebt = total
			item.DebtLimitRemain = item.DebtLimit - total
		}
		if item.ReceivablesLimit > 0 {
			var total = s.GetTotalReceivable(&item)
			item.TotalReceivable = total
			item.ReceivablesLimitRemain = item.ReceivablesLimit - total
		}

		profile, err := item.GetProfilePicture(s.ctx.DB)
		if err == nil {
			item.ProfilePicture = profile
		}

		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

// GetTotalDebt returns the total debt that a contact has.
//
// It takes in a contact and returns the total debt that the contact has.
// The total debt is calculated by summing up all the unpaid invoices of the contact.
// If the contact does not have any unpaid invoices, it returns 0.
func (s *ContactService) GetTotalDebt(contact *models.ContactModel) float64 {
	var total float64
	if err := s.ctx.DB.Model(&models.SalesModel{}).
		Where("document_type = ?", models.INVOICE).
		Where("contact_id = ?", contact.ID).
		Where("status IN (?)", []string{"POSTED", "FINISHED"}).
		Select("COALESCE(SUM(total - paid), 0)  as total").
		Scan(&total).Error; err != nil {
		return 0
	}
	return total
}

// GetTotalReceivable calculates the total receivables for a contact.
//
// It takes a contact model as input and returns the sum of all unpaid bills
// associated with the contact. The function queries the database for purchase
// orders with the document type 'BILL', belonging to the specified contact, and
// with a status of either 'POSTED' or 'FINISHED'. The total receivable amount
// is computed by summing up the difference between the total and paid amounts
// of these purchase orders. If no such records are found, it returns 0.

func (s *ContactService) GetTotalReceivable(contact *models.ContactModel) float64 {
	var total float64
	if err := s.ctx.DB.Model(&models.PurchaseOrderModel{}).
		Where("document_type = ?", models.BILL).
		Where("contact_id = ?", contact.ID).
		Where("status IN (?)", []string{"POSTED", "FINISHED"}).
		Select("COALESCE(SUM(total - paid), 0) as total").
		Scan(&total).Error; err != nil {
		return 0
	}
	return total
}

// Migrate migrates the database schema for the ContactService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the ContactModel schema.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.

func (s *ContactService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	return s.ctx.DB.AutoMigrate(&models.ContactModel{})
}

// DB returns the underlying database connection.
//
// The method returns the GORM database connection that is used by the service
// for CRUD (Create, Read, Update, Delete) operations.
func (s *ContactService) DB() *gorm.DB {
	return s.ctx.DB
}

// CountContactByTag returns a list of tags with the number of contacts associated
// with each tag, filtered by the given company ID.
//
// The returned list is grouped by tag ID, name, and color. The count of contacts
// associated with each tag is included in the response.
//
// If the query encounters an error, it will return that error. Otherwise, it will
// return the list of tags with contact counts.
func (s *ContactService) CountContactByTag(companyID string) ([]models.CountByTag, error) {
	var tag []models.CountByTag
	if err := s.ctx.DB.Model(&models.ContactModel{}).
		Joins("JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
		Joins("JOIN tags ON contact_tags.tag_model_id = tags.id").
		Select("tags.id, tags.name, tags.color, COUNT(*) as count").
		Where("tags.company_id = ?", companyID).
		Group("tags.id, tags.name, tags.color").
		Scan(&tag).Error; err != nil {
		return nil, err
	}
	return tag, nil
}

// GetContactByTagIDs returns a list of contacts associated with the given tag IDs.
//
// The returned list is filtered by the given tag IDs. If the query encounters an
// error, it will return that error. Otherwise, it will return the list of contacts.
func (s *ContactService) GetContactByTagIDs(tagIDs []string) ([]models.ContactModel, error) {
	var contacts []models.ContactModel
	if err := s.ctx.DB.Model(&models.ContactModel{}).
		Joins("JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
		Joins("JOIN tags ON contact_tags.tag_model_id = tags.id").
		Where("tags.id IN (?)", tagIDs).
		Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}
