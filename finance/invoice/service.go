package invoice

import (
	"net/http"

	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type InvoiceService struct {
	db *gorm.DB
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

func (s *InvoiceService) CreateInvoice(data *InvoiceModel) error {
	return s.db.Create(data).Error
}

func (s *InvoiceService) UpdateInvoice(id string, data *InvoiceModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *InvoiceService) DeleteInvoice(id string) error {
	return s.db.Where("id = ?", id).Delete(&InvoiceModel{}).Error
}

func (s *InvoiceService) GetInvoiceByID(id string) (*InvoiceModel, error) {
	var invoice InvoiceModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *InvoiceService) GetInvoiceByCode(code string) (*InvoiceModel, error) {
	var invoice InvoiceModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *InvoiceService) GetInvoices(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("invoices.description LIKE ? OR invoices.code LIKE ? ",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&InvoiceModel{})
	page := pg.With(stmt).Request(request).Response(&[]InvoiceModel{})
	return page, nil
}
