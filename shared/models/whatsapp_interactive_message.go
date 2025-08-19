package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

type WhatsappInteractiveMessage struct {
	shared.BaseModel
	MongoID     primitive.ObjectID `bson:"_id,omitempty" json:"object_id" gorm:"-"`
	Title       string             `json:"title" gorm:"type:varchar(255);not null" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Type        string             `json:"type" gorm:"type:varchar(255);not null" bson:"type"`
	RefID       *string            `json:"ref_id,omitempty" gorm:"index" bson:"refId"`
	RefType     *string            `json:"ref_type,omitempty" bson:"refType"`
	Data        json.RawMessage    `json:"data" gorm:"type:JSON" bson:"data"`
}

// BeforeCreate set object_id to MongoID
func (w *WhatsappInteractiveMessage) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		tx.Statement.SetColumn("id", utils.Uuid())
	}
	return nil
}
