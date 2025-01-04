package contact

import (
	"errors"
	"net/http"

	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ContactService struct {
	db            *gorm.DB
	SkipMigration bool
}

func NewContactService(db *gorm.DB, skipMigrate bool) *ContactService {
	var contactService = ContactService{db: db}
	if !skipMigrate {
		if err := contactService.Migrate(); err != nil {
			return nil
		}
	}
	return &contactService
}

// CreateContact membuat contact baru
func (s *ContactService) CreateContact(data *ContactModel) error {
	return s.db.Create(data).Error
}

// GetContactByID mengambil contact berdasarkan ID
func (s *ContactService) GetContactByID(id uint) (*ContactModel, error) {
	var contact ContactModel
	if err := s.db.First(&contact, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact not found")
		}
		return nil, err
	}
	return &contact, nil
}

// UpdateContactRoles mengupdate roles (is_customer, is_vendor, is_supplier) dari contact
func (s *ContactService) UpdateContactRoles(id uint, isCustomer, isVendor, isSupplier bool) (*ContactModel, error) {
	var contact ContactModel
	if err := s.db.First(&contact, id).Error; err != nil {
		return nil, err
	}

	contact.IsCustomer = isCustomer
	contact.IsVendor = isVendor
	contact.IsSupplier = isSupplier

	if err := s.db.Save(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetContactsByRole mengambil semua contact berdasarkan role (customer, vendor, supplier)
func (s *ContactService) GetContactsByRole(isCustomer, isVendor, isSupplier bool) ([]ContactModel, error) {
	var contacts []ContactModel
	query := s.db.Where("is_customer = ? AND is_vendor = ? AND is_supplier = ?", isCustomer, isVendor, isSupplier)
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// DeleteContact menghapus contact berdasarkan ID
func (s *ContactService) DeleteContact(id uint) error {
	if err := s.db.Delete(&ContactModel{}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("contact not found")
		}
		return err
	}
	return nil
}

// UpdateContact mengupdate informasi contact
func (s *ContactService) UpdateContact(id uint, data *ContactModel) (*ContactModel, error) {
	var contact ContactModel
	if err := s.db.First(&contact, id).Error; err != nil {
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

	if err := s.db.Save(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// GetContacts mengambil semua contact dengan pagination
func (s *ContactService) GetContacts(request http.Request, search string, isCustomer, isVendor, isSupplier *bool) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("contacts.name LIKE ? OR contacts.email LIKE ? OR contacts.phone LIKE ? OR contacts.address LIKE ?",
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
	stmt = stmt.Model(&ContactModel{})
	page := pg.With(stmt).Request(request).Response(&[]ContactModel{})
	return page, nil
}

func (s *ContactService) Migrate() error {
	if s.SkipMigration {
		return nil
	}
	return s.db.AutoMigrate(&ContactModel{})
}

func (s *ContactService) DB() *gorm.DB {
	return s.db
}
