package video

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type VideoService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewVideoService(db *gorm.DB, ctx *context.ERPContext) *VideoService {
	return &VideoService{
		db:  db,
		ctx: ctx,
	}
}
