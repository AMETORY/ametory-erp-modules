package project

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProjectService(ctx *context.ERPContext) *ProjectService {
	return &ProjectService{db: ctx.DB, ctx: ctx}
}

func (s *ProjectService) CreateProject(data *models.ProjectModel) error {
	return s.db.Omit(clause.Associations).Create(data).Error
}

func (s *ProjectService) UpdateProject(id string, data *models.ProjectModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProjectModel{}).Error
}

func (s *ProjectService) GetProjectByID(id string, memberID *string) (*models.ProjectModel, error) {
	var project models.ProjectModel
	db := s.db.Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" asc`).Preload("Tasks").Preload("Actions")
	}).Preload("Members.User")
	if memberID != nil {
		db = db.
			Joins("JOIN project_members ON project_members.project_model_id = projects.id").
			// Joins("JOIN members ON members.id = project_members.member_model_id").
			Where("project_members.member_model_id = ?", *memberID)
	}
	err := db.Where("id = ?", id).First(&project).Error
	return &project, err
}

func (s *ProjectService) GetColumnActionsByColumnID(id string) ([]models.ColumnAction, error) {
	var action []models.ColumnAction
	err := s.db.Where("column_id = ?", id).Find(&action).Error
	if err != nil {
		return []models.ColumnAction{}, err
	}
	return action, nil
}

func (s *ProjectService) GetProjects(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Columns").Preload("Members.User")
	if search != "" {
		stmt = stmt.Where("projects.description ILIKE ? OR projects.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.
			Joins("JOIN project_members ON project_members.project_model_id = projects.id").
			// Joins("JOIN members ON members.id = project_members.member_model_id").
			Where("project_members.member_model_id = ?", *memberID)
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("updated_at desc")
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ProjectModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProjectModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ProjectService) CreateColumn(data *models.ColumnModel) error {
	return s.db.Create(data).Error
}

func (s *ProjectService) UpdateColumn(id string, data *models.ColumnModel) error {
	return s.db.Where("id = ?", id).Omit(clause.Associations).Updates(data).Error
}

func (s *ProjectService) DeleteColumn(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ColumnModel{}).Error
}

func (s *ProjectService) GetColumnByID(id string) (*models.ColumnModel, error) {
	var invoice models.ColumnModel
	err := s.db.Where("id = ?", id).Preload("Actions").First(&invoice).Error
	return &invoice, err
}

func (s *ProjectService) GetColumns(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("columns.name ILIKE ?",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ColumnModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ColumnModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ProjectService) AddMemberToProject(projectID string, memberID string) error {
	return s.db.Table("project_members").Create(map[string]interface{}{
		"project_model_id": projectID,
		"member_model_id":  memberID,
	}).Error
}

func (s *ProjectService) GetMembersByProjectID(projectID string) ([]models.MemberModel, error) {
	var project models.ProjectModel
	err := s.db.Model(&models.ProjectModel{}).Where("id = ?", projectID).Preload("Members.User").Find(&project).Error
	return project.Members, err
}

func (s *ProjectService) AddActivity(projectID, memberID string, columnID, taskID *string, activityType string, notes *string) (*models.ProjectActivityModel, error) {
	var activity models.ProjectActivityModel = models.ProjectActivityModel{
		ProjectID:    projectID,
		MemberID:     memberID,
		TaskID:       taskID,
		ColumnID:     columnID,
		ActivityType: activityType,
		Notes:        notes,
	}

	if err := s.db.Create(&activity).Error; err != nil {
		return nil, err
	}
	// update project updatedAt
	if err := s.db.Model(&models.ProjectModel{}).Where("id = ?", projectID).Update("updated_at", time.Now()).Error; err != nil {
		return nil, err
	}

	return &activity, nil
}

func (s *ProjectService) GetRecentActivities(projectID string, limit int) ([]models.ProjectActivityModel, error) {
	var activities []models.ProjectActivityModel
	err := s.db.
		Preload("Project").Preload("Member.User").Preload("Column").Preload("Task").
		Where("project_id = ?", projectID).
		Order("activity_date desc").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}

func (s *ProjectService) CreateColumnAction(data *models.ColumnAction) error {
	return s.db.Create(data).Error
}

func (s *ProjectService) UpdateColumnAction(id string, data *models.ColumnAction) error {
	fmt.Println("UPDATE_COLUMN_ACTION", data)
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProjectService) DeleteColumnAction(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ColumnAction{}).Error
}

func (s *ProjectService) CheckIdleColumn() error {
	// Add logic to check for idle columns
	var idleColumns []models.ColumnAction
	err := s.db.Where("action_trigger = ?", "IDLE").Preload("Column.Tasks.CreatedBy").Find(&idleColumns).Error
	if err != nil {
		return err
	}

	for _, action := range idleColumns {

		if action.Action == "send_whatsapp_message" {
			for _, task := range action.Column.Tasks {
				var waSession models.WhatsappMessageSession
				if task.RefID != nil && *task.RefType == "whatsapp_session" {
					err := s.ctx.DB.Preload("Contact").First(&waSession, "id = ?", task.RefID).Error
					if err == nil {
						// utils.LogJson(waSession)
						actionData := map[string]any{}
						err := json.Unmarshal(*action.ActionData, &actionData)
						if err != nil {
							fmt.Println("ERROR UNMARSHAL", err)
							continue
						}
						now := time.Now()
						idleTime, ok := actionData["idle_time"].(float64)
						if !ok {
							fmt.Println("ERROR PARSING FLOAT", err)
							continue
						}
						idlePeriode, ok := actionData["idle_time_type"].(string)
						if !ok {
							fmt.Println("ERROR PARSING STRING", err)
							continue
						}
						readyToSend := false

						switch idlePeriode {
						case "days":
							fmt.Println(now.Sub(*task.UpdatedAt).Hours()/24, "HARI")
							if now.Sub(*task.UpdatedAt).Hours()/24 > idleTime {
								readyToSend = true
							}
						case "hours":
							fmt.Println(now.Sub(*task.UpdatedAt).Hours(), "JAM")
							if now.Sub(*task.UpdatedAt).Hours() > idleTime {
								readyToSend = true
							}
						case "minutes":
							fmt.Println(now.Sub(*task.UpdatedAt).Minutes(), "MENIT")
							if now.Sub(*task.UpdatedAt).Minutes() > idleTime {
								readyToSend = true
							}
						}

						if readyToSend {
							msg := parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string))
							_, err := sendWAMessage(s.ctx, waSession.JID, *waSession.Contact.Phone, msg)
							if err != nil {
								fmt.Println("ERROR SENDING MESSAGE", err)
								continue
							}
							task.LastActionTriggerAt = &now
							task.UpdatedAt = &now
							s.ctx.DB.Omit(clause.Associations).Save(&task)
						}
						// if waSession.Contact.Phone != nil {
						// 	msg := parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string))
						// 	_, err := sendWAMessage(s.ctx, waSession.JID, *waSession.Contact.Phone, msg)
						// 	if err != nil {
						// 		fmt.Println("ERROR SENDING MESSAGE", err)
						// 		continue
						// 	}

						// }
					}
				}
			}
		}
	}

	return nil
}

func parseTemplateID(msg string) *string {
	// Get UUID from string
	uuidRe := regexp.MustCompile(`@@\[([^\]]+)\]\(([^)]+)\)`)
	match := uuidRe.FindStringSubmatch(msg)
	if len(match) > 2 {
		msg = strings.ReplaceAll(msg, match[0], match[2])
		fmt.Println("MATCHED UUID", match[2])
		fmt.Println("MATCHED Msg", msg)
		return &msg
	}
	return nil

}

func parseMsgTemplate(contact models.ContactModel, member *models.MemberModel, msg string) string {
	re := regexp.MustCompile(`@\[([^\]]+)\]|\(\{\{([^}]+)\}\}\)`)

	// Replace
	result := re.ReplaceAllStringFunc(msg, func(s string) string {
		matches := re.FindStringSubmatch(s)
		re2 := regexp.MustCompile(`@\[([^\]]+)\]`)
		if re2.MatchString(s) {
			return ""
		}

		if matches[0] == "({{user}})" {
			return contact.Name
		}
		if matches[0] == "({{phone}})" {
			return *contact.Phone
		}
		if matches[0] == "({{agent}})" && member != nil {
			return member.User.FullName
		}
		return s // Kalau tidak ada datanya, biarkan
	})

	return result
}

func sendWAMessage(erpContext *context.ERPContext, jid, to, message string) (any, error) {
	replyData := whatsmeow_client.WaMessage{
		JID:     jid,
		Text:    message,
		To:      to,
		IsGroup: false,
	}
	// utils.LogJson(replyData)
	return erpContext.ThirdPartyServices["WA"].(*whatsmeow_client.WhatsmeowService).SendMessage(replyData)
}
