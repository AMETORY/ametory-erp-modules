package models

import (
	"strings"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileModel struct {
	shared.BaseModel
	FileName string     `gorm:"type:varchar(255)" json:"file_name"`
	MimeType string     `gorm:"type:varchar(255)" json:"mime_type"`
	Path     string     `gorm:"type:varchar(255)" json:"path"`
	Provider string     `gorm:"type:varchar(255);default:local" json:"provider"`
	URL      string     `gorm:"type:varchar(255)" json:"url"`
	RefID    string     `gorm:"type:char(36);index" json:"ref_id,omitempty"`
	RefType  string     `gorm:"type:varchar(255);index" json:"ref_type,omitempty"`
	SkipSave bool       `gorm:"-" json:"-"`
	UserID   *string    `gorm:"type:char(36);index" json:"user_id,omitempty"`
	User     *UserModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
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

func GetThumbnail(files []FileModel) (*FileModel, []FileModel) {
	restFiles := []FileModel{}
	var thumbnail *FileModel
	for _, v := range files {
		if strings.HasPrefix(v.MimeType, "image/") && thumbnail == nil {
			thumbnail = &v
		} else {
			restFiles = append(restFiles, v)
		}
	}
	return thumbnail, restFiles
}
