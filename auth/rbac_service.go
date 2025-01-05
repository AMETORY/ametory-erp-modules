package auth

import (
	"errors"

	"gorm.io/gorm"
)

type RBACService struct {
	db *gorm.DB
}

func NewRBACService(db *gorm.DB) *RBACService {
	return &RBACService{db: db}
}

// AssignRoleToUser menetapkan peran ke pengguna
func (s *RBACService) AssignRoleToUser(userID string, roleName string) error {
	var user UserModel
	var role RoleModel

	// Cari pengguna
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Cari peran
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return errors.New("role not found")
	}

	// Tetapkan peran ke pengguna
	if err := s.db.Model(&user).Association("Roles").Append(&role); err != nil {
		return err
	}

	return nil
}

// AssignPermissionToRole menetapkan izin ke peran
func (s *RBACService) AssignPermissionToRole(roleName string, permissionName string) error {
	var role RoleModel
	var permission PermissionModel

	// Cari peran
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return errors.New("role not found")
	}

	// Cari izin
	if err := s.db.Where("name = ?", permissionName).First(&permission).Error; err != nil {
		return errors.New("permission not found")
	}

	// Tetapkan izin ke peran
	if err := s.db.Model(&role).Association("Permissions").Append(&permission); err != nil {
		return err
	}

	return nil
}

// CheckPermission memeriksa apakah pengguna memiliki izin tertentu
func (s *RBACService) CheckPermission(userID string, permissionName string) (bool, error) {
	var user UserModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return false, errors.New("user not found")
	}

	// Periksa izin
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true, nil
			}
		}
	}

	return false, nil
}
