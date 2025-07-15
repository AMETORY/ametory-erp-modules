package cooperative_member

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type CooperativeMemberService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewCooperativeMemberService creates a new instance of CooperativeMemberService.
//
// The service provides methods for creating, updating, deleting, and retrieving
// cooperative members.
//
// The service requires a Gorm database instance and an ERP context.
func NewCooperativeMemberService(db *gorm.DB, ctx *context.ERPContext) *CooperativeMemberService {
	return &CooperativeMemberService{
		db:  db,
		ctx: ctx,
	}
}

// CreateMember adds a new cooperative member to the database.
// It takes a CooperativeMemberModel as input and returns an error if the operation fails.

func (s *CooperativeMemberService) CreateMember(data *models.CooperativeMemberModel) error {
	return s.db.Create(data).Error
}

// UpdateMember updates a cooperative member in the database.
//
// It takes a string id and a pointer to a CooperativeMemberModel as parameters and
// returns an error. It uses the gorm.DB connection to update a record in the
// cooperative_members table.
func (s *CooperativeMemberService) UpdateMember(id string, data *models.CooperativeMemberModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteMember deletes a cooperative member from the database.
//
// It takes a string id as parameter and returns an error. It uses the gorm.DB
// connection to delete a record in the cooperative_members table.
func (s *CooperativeMemberService) DeleteMember(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.CooperativeMemberModel{}).Error
}

// GetMemberByID retrieves a cooperative member by its ID.
//
// It takes a string id as parameter and returns a pointer to a CooperativeMemberModel
// and an error. The function uses GORM to retrieve the member data from the
// cooperative_members table. If the operation fails, an error is returned.
func (s *CooperativeMemberService) GetMemberByID(id string) (*models.CooperativeMemberModel, error) {
	var invoice models.CooperativeMemberModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetMembers retrieves a paginated list of cooperative members from the database.
//
// It takes an HTTP request and a search query string as parameters. The search query
// is applied to the cooperative member's name, email, and phone fields. If a company ID
// is present in the request header, the result is filtered by the company ID. The
// function uses pagination to manage the result set and includes any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of CooperativeMemberModel and an error if the
// operation fails.
func (s *CooperativeMemberService) GetMembers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("User").Preload("Role")
	if search != "" {
		stmt = stmt.
			Joins("LEFT JOIN users ON users.id = members.user_id")
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)

	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.CooperativeMemberModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.CooperativeMemberModel{})
	page.Page = page.Page + 1
	return page, nil
}

// InviteMember creates a new invitation for a cooperative member in the database.
//
// It takes a pointer to a MemberInvitationModel as a parameter and returns the
// generated token and an error. The function generates a random token if one is not
// provided in the passed model. It uses the gorm.DB connection to create a new
// record in the member_invitations table. If the operation fails, an error is
// returned.
func (s *CooperativeMemberService) InviteMember(data *models.MemberInvitationModel) (string, error) {
	if data.Token == "" {
		data.Token = utils.RandString(32, false)
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return "", err
	}
	return data.Token, nil
}

// GetInvitedMembers retrieves a paginated list of invited members from the database.
//
// It takes an HTTP request and a search query string as parameters. The search query
// is applied to the invited member's full name field. If a company ID is present in the
// request header, the result is filtered by the company ID. The function uses
// pagination to manage the result set and includes any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of MemberInvitationModel and an error if the
// operation fails.
func (s *CooperativeMemberService) GetInvitedMembers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Inviter").Preload("Role")

	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("full_name ILIKE ?",
			"%"+search+"%",
		)
	}

	utils.FixRequest(&request)
	stmt = stmt.Model(&models.MemberInvitationModel{})
	page := pg.With(stmt).Request(request).Response(&[]models.MemberInvitationModel{})
	utils.FixRequest(&request)
	page.Page = page.Page + 1
	return page, nil
}

// AcceptMemberInvitation accepts a member invitation and creates a new cooperative
// member in the database.
//
// It takes a token and a user ID as parameters. The token is used to find the
// corresponding invitation in the database. If the invitation is valid, a new
// cooperative member is created with the provided user ID and the role ID
// specified in the invitation. The function also deletes the invitation from the
// database. If the operation fails, an error is returned.
func (s *CooperativeMemberService) AcceptMemberInvitation(token string, userID string) error {
	var invitation models.MemberInvitationModel
	if err := s.db.Where("token = ?", token).First(&invitation).Error; err != nil {
		return err
	}
	data := models.CooperativeMemberModel{
		ConnectedTo: &userID,
		CompanyID:   invitation.CompanyID,
		RoleID:      invitation.RoleID,
		Status:      "ACTIVE",
		Active:      true,
	}
	err := s.db.Create(&data).Error
	if err != nil {
		return err
	}

	err = s.db.Delete(&invitation).Error
	if err != nil {
		return err
	}

	// add role to user

	return nil
}

// DeleteInvitation deletes a member invitation in the database.
//
// It takes an ID as parameter and returns an error if the invitation does not exist or if
// there is an issue deleting the invitation from the database.
func (s *CooperativeMemberService) DeleteInvitation(id string) error {
	var invitation models.MemberInvitationModel
	if err := s.db.Where("id = ?", id).First(&invitation).Error; err != nil {
		return err
	}

	err := s.db.Delete(&invitation).Error
	if err != nil {
		return err
	}

	return nil
}

// ApproveMemberByID approves a cooperative member by its ID.
//
// It takes a member ID and a user ID as parameters. The member ID is used to find the
// corresponding member in the database. If the member is found, its status is set to
// "APPROVED", and the approved by and approved at fields are updated with the
// provided user ID and the current time. The function uses the gorm.DB connection to
// update the member in the database. If the operation fails, an error is returned.
func (s *CooperativeMemberService) ApproveMemberByID(id string, userID string) error {
	var member models.CooperativeMemberModel
	if err := s.db.Where("id = ?", id).First(&member).Error; err != nil {
		return err
	}
	now := time.Now()
	member.Status = "APPROVED"
	member.ApprovedBy = &userID
	member.ApprovedAt = &now

	err := s.db.Save(&member).Error
	if err != nil {
		return err
	}
	return nil
}
