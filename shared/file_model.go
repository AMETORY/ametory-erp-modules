package shared

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileModel struct {
	utils.BaseModel
	FileName string `gorm:"type:varchar(255)"`
	MimeType string `gorm:"type:varchar(255)"`
	Path     string `gorm:"type:varchar(255)"`
	Provider string `gorm:"type:varchar(255);default:local"`
	URL      string `gorm:"type:varchar(255)"`
	RefID    string `gorm:"type:char(36);index"`
	RefType  string `gorm:"type:varchar(255);index"`
}

func (FileModel) TableName() string {
	return "files"
}
func (f *FileModel) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&FileModel{})
}
