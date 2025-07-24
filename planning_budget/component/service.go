package component

import "github.com/AMETORY/ametory-erp-modules/context"

type ComponentService struct {
	ctx *context.ERPContext
}

func NewComponentService(ctx *context.ERPContext) *ComponentService {
	return &ComponentService{ctx: ctx}
}
