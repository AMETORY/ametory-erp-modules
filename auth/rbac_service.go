package auth

import (
	"errors"
	"net/http"

	"github.com/morkid/paginate"
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
func (s *RBACService) CheckPermission(userID string, permissionNames []string) (bool, error) {
	var user UserModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_admin = ?", false).Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Where("name IN (?)", permissionNames)
		})
	}).First(&user, userID).Error; err != nil {
		return false, errors.New("user not found")
	}

	// Periksa izin
	for _, roleName := range permissionNames {
		for _, role := range user.Roles {
			if role.IsSuperAdmin {
				return true, nil
			}
			for _, permission := range role.Permissions {
				if permission.Name == roleName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CheckPermission memeriksa apakah pengguna memiliki izin tertentu
func (s *RBACService) CheckAdminPermission(adminID string, permissionNames []string) (bool, error) {
	var admin AdminModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_admin = ?", true).
			Preload("Permissions", func(db *gorm.DB) *gorm.DB {
				return db.Where("name IN (?)", permissionNames)
			})
	}).First(&admin, "id = ?", adminID).Error; err != nil {
		return false, errors.New("admin not found")
	}

	// Periksa izin
	for _, roleName := range permissionNames {
		for _, role := range admin.Roles {
			if role.IsSuperAdmin {
				return true, nil
			}
			for _, permission := range role.Permissions {
				if permission.Name == roleName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CreateRole membuat peran baru
func (s *RBACService) CreateRole(name string, isAdmin bool, isSuperAdmin bool, companyID *string) (*RoleModel, error) {
	role := RoleModel{
		Name:         name,
		IsAdmin:      isAdmin,
		IsSuperAdmin: isSuperAdmin,
		CompanyID:    companyID,
	}

	if err := s.db.Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// GetAllRoles mengambil semua peran
func (s *RBACService) GetAllRoles(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("roles.name LIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	stmt = stmt.Model(&RoleModel{})
	page := pg.With(stmt).Request(request).Response(&[]RoleModel{})
	return page, nil
}

// GetRoleByID mengambil peran berdasarkan ID
func (s *RBACService) GetRoleByID(roleID string) (*RoleModel, error) {
	var role RoleModel
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, errors.New("role not found")
	}
	return &role, nil
}

// UpdateRole memperbarui informasi peran berdasarkan ID
func (s *RBACService) UpdateRole(roleID string, name string, isAdmin bool, isSuperAdmin bool) (*RoleModel, error) {
	var role RoleModel
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, errors.New("role not found")
	}

	role.Name = name
	role.IsAdmin = isAdmin
	role.IsSuperAdmin = isSuperAdmin

	if err := s.db.Save(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole menghapus peran berdasarkan ID
func (s *RBACService) DeleteRole(roleID string) error {
	if err := s.db.Delete(&RoleModel{}, roleID).Error; err != nil {
		return err
	}
	return nil
}
