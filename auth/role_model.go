package auth

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleModel struct {
	shared.BaseModel
	Name         string            `gorm:"unique;not null" json:"name"`
	Permissions  []PermissionModel `gorm:"many2many:role_permissions;" json:"permissions"`
	CompanyID    *string           `json:"company_id"`
	IsAdmin      bool              `json:"is_admin"`
	IsMerchant   bool              `json:"is_merchant"`
	IsSuperAdmin bool              `json:"is_super_admin"`
}

// PermissionModel adalah model database untuk izin
type PermissionModel struct {
	shared.BaseModel
	Name     string `gorm:"unique;not null" json:"name"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

func (RoleModel) TableName() string {
	return "roles"
}

func (PermissionModel) TableName() string {
	return "permissions"
}

func (r *RoleModel) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return
}
func (r *PermissionModel) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return
}

var (
	cruds    = []string{"create", "read", "update", "delete"}
	services = map[string][]map[string][]string{
		"auth":    {{"user": cruds, "admin": cruds, "rbac": cruds}},
		"finance": {{"account": cruds, "transaction": cruds}},
		"inventory": {
			{"brand": cruds},
			{"product_category": cruds},
			{"product": cruds},
			{"master_product": cruds},
			{"warehouse": cruds},
			{"stock_movement": cruds},
			{"purchase": cruds},
			{"procurement": cruds},
		},
		"contact": {
			{"customer": cruds},
			{"vendor": cruds},
			{"supplier": cruds},
		},
		"order": {
			{"sales": cruds},
			{"pos": cruds},
			{"merchant": cruds},
		},
		"distribution": {
			{"distributor": cruds},
			{"offering": cruds},
			{"order_request": append(cruds, "approval")},
		},
	}
)

func GeneratePermissions() []PermissionModel {
	var permissions []PermissionModel

	for service, modules := range services {
		for _, module := range modules {
			for key, actions := range module {
				for _, action := range actions {
					permissions = append(permissions, PermissionModel{Name: service + ":" + key + ":" + action})
				}

			}
		}
	}

	return permissions
}
