package contact

import (
	"errors"
	"net/http"

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

// CreateContact membuat contact baru
func (s *ContactService) CreateContact(data *models.ContactModel) error {
	return s.ctx.DB.Create(data).Error
}

// CreateContactFromUser membuat contact baru dari user jika tidak ada contact yang sama dengan email user
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

// GetContactByID mengambil contact berdasarkan ID
func (s *ContactService) GetContactByID(id string) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.First(&contact, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}
	return &contact, nil
}

// UpdateContactRoles mengupdate roles (is_customer, is_vendor, is_supplier) dari contact
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

// GetContactsByRole mengambil semua contact berdasarkan role (customer, vendor, supplier)
func (s *ContactService) GetContactsByRole(isCustomer, isVendor, isSupplier bool) ([]models.ContactModel, error) {
	var contacts []models.ContactModel
	query := s.ctx.DB.Where("is_customer = ? AND is_vendor = ? AND is_supplier = ?", isCustomer, isVendor, isSupplier)
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// DeleteContact menghapus contact berdasarkan ID
func (s *ContactService) DeleteContact(id string) error {
	if err := s.ctx.DB.Delete(&models.ContactModel{}, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contact not found")
		}
		return err
	}
	return nil
}

// UpdateContact mengupdate informasi contact
func (s *ContactService) UpdateContact(id string, data *models.ContactModel) (*models.ContactModel, error) {
	var contact models.ContactModel
	if err := s.ctx.DB.First(&contact, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}

	// Update fields
	contact.Name = data.Name
	contact.Email = data.Email
	contact.Code = data.Code
	contact.Phone = data.Phone
	contact.Address = data.Address
	contact.ContactPerson = data.ContactPerson
	contact.ContactPersonPosition = data.ContactPersonPosition
	contact.IsCustomer = data.IsCustomer
	contact.IsVendor = data.IsVendor
	contact.IsSupplier = data.IsSupplier

	if err := s.ctx.DB.Save(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetContacts mengambil semua contact dengan pagination
func (s *ContactService) GetContacts(request http.Request, search string, isCustomer, isVendor, isSupplier *bool) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB
	if search != "" {
		stmt = stmt.Where("contacts.name ILIKE ? OR contacts.email ILIKE ? OR contacts.phone ILIKE ? OR contacts.address ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if isCustomer != nil {
		stmt = stmt.Where("is_customer = ?", isCustomer)
	}
	if isVendor != nil {
		stmt = stmt.Where("is_vendor = ?", isVendor)
	}
	if isSupplier != nil {
		stmt = stmt.Where("is_supplier = ?", isSupplier)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.ContactModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ContactModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ContactService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	return s.ctx.DB.AutoMigrate(&models.ContactModel{})
}

func (s *ContactService) DB() *gorm.DB {
	return s.ctx.DB
}
