package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type RBACService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
	mode       string
}

func NewRBACService(erpContext *context.ERPContext) *RBACService {
	return &RBACService{erpContext: erpContext, db: erpContext.DB}
}

func (s *RBACService) SetMode(mode string) {
	s.mode = mode
}

// AssignRoleToUser menetapkan peran ke pengguna
func (s *RBACService) AssignRoleToUser(userID string, roleName string) error {
	var user models.UserModel
	var role models.RoleModel

	// Cari pengguna
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
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

// AssignRoleToAdmin menetapkan peran ke admin
func (s *RBACService) AssignRoleToAdmin(adminID string, roleName string) error {
	var admin models.AdminModel
	var role models.RoleModel

	// Cari admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}

	// Cari peran
	if err := s.db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return errors.New("role not found")
	}

	// Tetapkan peran ke admin
	if err := s.db.Model(&admin).Association("Roles").Append(&role); err != nil {
		return err
	}

	return nil
}

// AssignPermissionToRole menetapkan izin ke peran
func (s *RBACService) AssignPermissionToRole(roleName string, permissionName string) error {
	var role models.RoleModel
	var permission models.PermissionModel

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
func (s *RBACService) AssignPermissionToRoleID(id string, permissionName string) error {
	var role models.RoleModel
	var permission models.PermissionModel

	// Cari peran
	if err := s.db.Where("id = ?", id).First(&role).Error; err != nil {
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
	var user models.UserModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_admin = ?", false).Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Where("name IN (?)", permissionNames)
		})
	}).First(&user, "id = ?", userID).Error; err != nil {
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

	return false, fmt.Errorf("permissions %s not found", strings.Join(permissionNames, ", "))
}
func (s *RBACService) CheckPermissionWithCompanyID(userID, companyID string, permissionNames []string) (bool, error) {
	var user models.UserModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Where("company_id = ?", companyID).Where("is_admin = ?", false).Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Where("name IN (?)", permissionNames)
		})
	}).First(&user, "id = ?", userID).Error; err != nil {
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
	var admin models.AdminModel

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
func (s *RBACService) CheckSuperAdminPermission(adminID string) (bool, error) {
	var admin models.AdminModel

	// Cari pengguna beserta peran dan izin
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_admin = ?", true)
	}).First(&admin, "id = ?", adminID).Error; err != nil {
		return false, errors.New("admin not found")
	}

	// Periksa izin
	for _, role := range admin.Roles {
		if role.IsSuperAdmin {
			return true, nil
		}
	}

	return false, nil
}

// AddPermissionsToRole menambahkan beberapa izin ke peran
func (s *RBACService) AddPermissionsToRole(roleID string, permissionNames []string) error {
	var role models.RoleModel
	var permissions []models.PermissionModel

	// Cari peran
	if err := s.db.First(&role, "id = ?", roleID).Error; err != nil {
		return errors.New("role not found")
	}

	// Cari izin
	if err := s.db.Where("name IN (?)", permissionNames).Find(&permissions).Error; err != nil {
		return errors.New("permissions not found")
	}

	// Tetapkan izin ke peran
	if err := s.db.Model(&role).Association("Permissions").Append(&permissions); err != nil {
		return err
	}

	return nil
}

// GetRoleFromUser mengambil peran berdasarkan user
func (s *RBACService) GetRoleFromUser(userID string) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.db.
		Joins("JOIN user_roles ON user_roles.role_model_id = roles.id").
		Joins("JOIN users ON users.id = user_roles.user_model_id").
		Where("users.id = ?", userID).
		First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName mengambil peran berdasarkan nama
func (s *RBACService) GetRoleByName(name string, isAdmin bool) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.db.Where("name = ? AND is_admin = ?", name, isAdmin).First(&role).Error; err != nil {
		return nil, errors.New("role not found")
	}
	return &role, nil
}

// CheckRoleExistsByName memeriksa apakah peran dengan nama tertentu ada
func (s *RBACService) CheckRoleExistsByName(name string, isAdmin bool) (bool, error) {
	var count int64
	if err := s.db.Model(&models.RoleModel{}).
		Where("name = ? AND is_admin = ?", name, isAdmin).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// CreateRole membuat peran baru
func (s *RBACService) CreateRole(name string, isAdmin, isSuperAdmin, isMerchant bool, companyID *string) (*models.RoleModel, error) {
	role := models.RoleModel{
		Name:         name,
		IsAdmin:      isAdmin,
		IsSuperAdmin: isSuperAdmin,
		IsMerchant:   isMerchant,
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
	stmt := s.db.Preload("Permissions")
	if search != "" {
		stmt = stmt.Where("roles.name ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("is_admin") != "" {
		stmt = stmt.Where("is_admin = ?", request.URL.Query().Get("is_admin"))
	}
	if request.URL.Query().Get("is_merchant") != "" {
		stmt = stmt.Where("is_merchant = ?", request.URL.Query().Get("is_merchant"))
	}
	if request.URL.Query().Get("is_owner") != "" {
		stmt = stmt.Where("is_owner = ?", request.URL.Query().Get("is_owner"))
	}
	if request.URL.Query().Get("is_super_admin") != "" {
		stmt = stmt.Where("is_super_admin = ?", request.URL.Query().Get("is_super_admin"))
	}

	if s.mode == "user" {
		stmt = stmt.Where("is_admin = ?", false)
	}

	stmt = stmt.Model(&models.RoleModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.RoleModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.RoleModel)
	newItems := make([]models.RoleModel, 0)

	for _, v := range *items {
		if v.IsSuperAdmin {
			var Permissions []models.PermissionModel
			s.db.Find(&Permissions)
			v.Permissions = Permissions
		}
		newItems = append(newItems, v)

	}
	page.Items = &newItems
	return page, nil
}

// GetRoleByID mengambil peran berdasarkan ID
func (s *RBACService) GetRoleByID(roleID string) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.db.First(&role, "id = ?", roleID).Error; err != nil {
		return nil, errors.New("role not found")
	}
	return &role, nil
}

// UpdateRole memperbarui informasi peran berdasarkan ID
func (s *RBACService) UpdateRole(roleID, name string, isAdmin, isSuperAdmin, isMerchant, isOwner bool) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.db.First(&role, "id = ?", roleID).Error; err != nil {
		return nil, errors.New("role not found")
	}

	role.Name = name
	role.IsAdmin = isAdmin
	role.IsSuperAdmin = isSuperAdmin
	role.IsMerchant = isMerchant
	role.IsOwner = isOwner

	if err := s.db.Save(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole menghapus peran berdasarkan ID
func (s *RBACService) DeleteRole(roleID string) error {
	if err := s.db.Delete(&models.RoleModel{}, "id = ?", roleID).Error; err != nil {
		return err
	}
	return nil
}

// GetAllPermissions mengambil semua izin
func (s *RBACService) GetAllPermissions() ([]models.PermissionModel, error) {
	var permissions []models.PermissionModel
	if err := s.db.Where("is_active = ?", true).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
