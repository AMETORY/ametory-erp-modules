package auth

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleModel struct {
	utils.BaseModel
	Name         string            `gorm:"unique;not null"`
	Permissions  []PermissionModel `gorm:"many2many:role_permissions;"`
	IsAdmin      bool
	IsSuperAdmin bool
}

// PermissionModel adalah model database untuk izin
type PermissionModel struct {
	utils.BaseModel
	Name string `gorm:"unique;not null"`
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
	services = map[string][]string{
		"auth":      {"user", "admin", "rbac"},
		"finance":   {"account", "transaction"},
		"inventory": {"brand", "product_category", "product", "master_product", "warehouse", "stock_movement", "purchase", "procurement"},
		"contact":   {"customer", "vendor", "supplier", "distributor"},
		"order":     {"sales", "pos"},
	}
)

func GeneratePermissions() []PermissionModel {
	var permissions []PermissionModel
	cruds := []string{"create", "read", "update", "delete"}
	for _, crudAction := range cruds {
		for service, actions := range services {
			for _, action := range actions {
				permissions = append(permissions, PermissionModel{Name: service + ":" + action + ":" + crudAction})
			}
		}
	}

	return permissions
}
