package shared

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileModel struct {
	BaseModel
	FileName string `gorm:"type:varchar(255)" json:"file_name"`
	MimeType string `gorm:"type:varchar(255)" json:"mime_type"`
	Path     string `gorm:"type:varchar(255)" json:"path"`
	Provider string `gorm:"type:varchar(255);default:local" json:"provider"`
	URL      string `gorm:"type:varchar(255)" json:"url"`
	RefID    string `gorm:"type:char(36);index" json:"-"`
	RefType  string `gorm:"type:varchar(255);index" json:"-"`
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
