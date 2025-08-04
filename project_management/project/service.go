package project

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectService struct {
	db                          *gorm.DB
	ctx                         *context.ERPContext
	whatsmeowService            *whatsmeow_client.WhatsmeowService
	customerRelationshipService *customer_relationship.CustomerRelationshipService
}

// NewProjectService creates a new instance of ProjectService with the given ERP context.
//
// It fetches the whatsmeow service from the context and panics if it is not found.
// It also fetches the CustomerRelationshipService from the context, but won't panic if it is not found.
// It then returns the newly created ProjectService with the given database connection, context, whatsmeow service and customer relationship service.
func NewProjectService(ctx *context.ERPContext) *ProjectService {
	whatsmeowService, ok := ctx.ThirdPartyServices["WA"].(*whatsmeow_client.WhatsmeowService)
	if !ok {
		panic("ThirdPartyServices is not instance of whatsmeow_client.WhatsmeowService")
	}
	var customerRelationshipService *customer_relationship.CustomerRelationshipService
	customerRelationshipSrv, ok := ctx.CustomerRelationshipService.(*customer_relationship.CustomerRelationshipService)
	if ok {
		customerRelationshipService = customerRelationshipSrv
	}

	return &ProjectService{db: ctx.DB, ctx: ctx, whatsmeowService: whatsmeowService, customerRelationshipService: customerRelationshipService}
}

// CreateProject creates a new project with the given data.
//
// It uses the Omit clause to avoid creating associations.
func (s *ProjectService) CreateProject(data *models.ProjectModel) error {
	return s.db.Omit(clause.Associations).Create(data).Error
}

// UpdateProject updates the project with the given ID to the given data.
//
// It uses the Updates method to update the columns of the project.
func (s *ProjectService) UpdateProject(id string, data *models.ProjectModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteProject deletes the project with the given ID.
//
// It uses the Delete method to delete the project.
func (s *ProjectService) DeleteProject(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProjectModel{}).Error
}

// GetProjectByID retrieves a project by its ID, optionally filtering by a member ID.
//
// It takes a project ID and an optional member ID as parameters. If a member ID is provided,
// the project is filtered to include only those where the given member is associated.
// The function preloads associated Columns, Tasks, Actions, and Members.User data.
//
// Returns a pointer to a ProjectModel and an error, if any.
func (s *ProjectService) GetProjectByID(id string, memberID *string) (*models.ProjectModel, error) {
	var project models.ProjectModel
	db := s.db.Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" asc`).Preload("Tasks").Preload("Actions")
	}).Preload("Members.User")
	if memberID != nil {
		db = db.
			Joins("JOIN project_members ON project_members.project_model_id = projects.id").
			Where("project_members.member_model_id = ?", *memberID)
	}
	err := db.Where("id = ?", id).First(&project).Error
	return &project, err
}

// GetColumnActionsByColumnID retrieves actions associated with a specific column ID.
//
// It takes a column ID as a parameter and returns a slice of ColumnAction and an error, if any.
// The function preloads the associated Column data, selecting only the ID and name fields.
func (s *ProjectService) GetColumnActionsByColumnID(id string) ([]models.ColumnAction, error) {
	var action []models.ColumnAction
	err := s.db.Where("column_id = ?", id).Preload("Column", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name") }).Find(&action).Error
	if err != nil {
		return []models.ColumnAction{}, err
	}
	return action, nil
}

// GetProjects retrieves a paginated list of projects, optionally filtering by search and member ID.
//
// It takes an HTTP request, a search string, and an optional member ID. The search string is applied to
// the project's name and description fields. If a company ID is present in the request header, the result is
// filtered by the company ID. If a member ID is provided, the projects are filtered to include only those
// associated with the given member. The function supports ordering and pagination.
//
// Returns a paginated page of ProjectModel and an error, if any.
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

// CreateColumn creates a new column with the given data.
//
// It uses the Omit clause to avoid creating associations.
func (s *ProjectService) CreateColumn(data *models.ColumnModel) error {
	return s.db.Create(data).Error
}

// UpdateColumn updates the column with the given ID to the given data.
//
// It uses the Updates method to update the columns of the column.
func (s *ProjectService) UpdateColumn(id string, data *models.ColumnModel) error {
	return s.db.Where("id = ?", id).Omit(clause.Associations).Updates(data).Error
}

// DeleteColumn deletes the column with the given ID.
//
// It uses the Delete method to delete the column.
func (s *ProjectService) DeleteColumn(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ColumnModel{}).Error
}

// GetColumnByID retrieves a column by its ID, optionally filtering by a project ID.
//
// It takes a column ID as a parameter and returns a pointer to a ColumnModel and an error, if any.
// The function preloads associated Actions, Project, and Column data.
func (s *ProjectService) GetColumnByID(id string) (*models.ColumnModel, error) {
	var column models.ColumnModel
	err := s.db.Where("id = ?", id).Preload("Actions").Preload("Project").First(&column).Error
	return &column, err
}

// GetColumns retrieves a paginated list of columns, optionally filtering by search and project ID.
//
// It takes an HTTP request, a search string, and an optional project ID. The search string is applied to
// the column's name field. If a project ID is present in the request header, the result is
// filtered by the project ID. The function supports ordering and pagination.
//
// Returns a paginated page of ColumnModel and an error, if any.
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

// AddMemberToProject adds a member to a project with the given ID and member ID.
//
// It creates a new project member entry in the project_members table.
func (s *ProjectService) AddMemberToProject(projectID string, memberID string) error {
	return s.db.Table("project_members").Create(map[string]interface{}{
		"project_model_id": projectID,
		"member_model_id":  memberID,
	}).Error
}

// GetMembersByProjectID retrieves a list of members associated with a project with the given ID.
//
// It takes a project ID as a parameter and returns a slice of MemberModel and an error, if any.
// The function preloads associated User data.
func (s *ProjectService) GetMembersByProjectID(projectID string) ([]models.MemberModel, error) {
	var project models.ProjectModel
	err := s.db.Model(&models.ProjectModel{}).Where("id = ?", projectID).Preload("Members.User").Find(&project).Error
	return project.Members, err
}

// AddActivity adds an activity to a project with the given ID.
//
// It takes a project ID, a member ID, a column ID, a task ID, an activity type, and an optional notes string as parameters.
// It creates a new project activity entry in the project_activities table.
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

// GetRecentActivities retrieves a list of recent activities associated with a project with the given ID.
//
// It takes a project ID and a limit as parameters and returns a slice of ProjectActivityModel and an error, if any.
// The function preloads associated Project, Member, Column, Task, and User data.
// The function supports ordering by the activity date in descending order.
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

// CreateColumnAction creates a new column action with the given data.
//
// It uses the Omit clause to avoid creating associations.
func (s *ProjectService) CreateColumnAction(data *models.ColumnAction) error {
	return s.db.Create(data).Error
}

// UpdateColumnAction updates the column action with the given ID to the given data.
//
// It uses the Updates method to update the columns of the column action.
func (s *ProjectService) UpdateColumnAction(id string, data *models.ColumnAction) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteColumnAction deletes the column action with the given ID.
//
// It uses the Delete method to delete the column action.
func (s *ProjectService) DeleteColumnAction(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ColumnAction{}).Error
}

// CheckIdleColumn checks for idle columns and sends a message to the column's tasks if the idle time has been reached.
//
// It takes a callback function as a parameter, which is called with the scheduled message data.
//
// The function queries for all column actions with the action trigger set to "IDLE" and preloads the associated Column, Tasks, and User data.
// For each column action, it checks if the action is set to "send_whatsapp_message" and if the task's ref type is set to "whatsapp_session".
// If so, it checks if the task's last action trigger at is older than the idle time and if the action status is set to "READY".
// If both conditions are met, it sets the action status to "WAITING", updates the task, and calls the callback function with the scheduled message data.
// The callback function is expected to send the message to the task's contact phone number.
func (s *ProjectService) CheckIdleColumn(callback func(models.ScheduledMessage)) error {
	// Add logic to check for idle columns
	var idleColumns []models.ColumnAction
	err := s.db.Where("action_trigger = ?", "IDLE").Preload("Column.Tasks.CreatedBy").Find(&idleColumns).Error
	if err != nil {
		return err
	}

	for _, action := range idleColumns {

		if action.Action == "send_whatsapp_message" {
			log.Println("READY TO GET TASK FROM", action.Name)
			for _, task := range action.Column.Tasks {
				var waSession models.WhatsappMessageSession
				if task.RefID != nil && *task.RefType == "whatsapp_session" {
					err := s.ctx.DB.Preload("Contact").First(&waSession, "id = ?", task.RefID).Error
					if err == nil {
						log.Println("READY TO EXCUTE IDLE TASK", waSession.Contact.Name)
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

						updatedAt := *task.UpdatedAt
						if task.LastActionTriggerAt != nil {
							updatedAt = *task.LastActionTriggerAt
						}

						switch idlePeriode {
						case "days":
							fmt.Println(now.Sub(updatedAt).Hours()/24, "HARI")
							if now.Sub(updatedAt).Hours()/24 > idleTime && action.ActionStatus == "READY" {
								readyToSend = true
							}
						case "hours":
							fmt.Println(now.Sub(updatedAt).Hours(), "JAM", action.ActionStatus)

							if now.Sub(updatedAt).Hours() > idleTime && action.ActionStatus == "READY" {
								readyToSend = true
							}
						case "minutes":
							fmt.Println(now.Sub(updatedAt).Minutes(), "MENIT")
							if now.Sub(updatedAt).Minutes() > idleTime && action.ActionStatus == "READY" {
								readyToSend = true
							}
						}

						if readyToSend {
							if action.ActionHour != nil {
								parsedTime, err := time.Parse("15:04", *action.ActionHour)
								if err != nil {
									fmt.Println("ERROR PARSING TIME", err)
									continue
								}
								// nowTime := time.Now().In(parsedTime.Location())
								delay := time.Duration(parsedTime.Hour()-time.Now().Hour())*time.Hour +
									time.Duration(parsedTime.Minute()-time.Now().Minute())*time.Minute
								action.ActionStatus = "WAITING"
								s.ctx.DB.Omit(clause.Associations).Save(&action)

								log.Println("DELAY TO EXCUTE NOW", time.Now())
								log.Println("DELAY TO EXCUTE IDLE TASK", delay)
								fmt.Println("DELAY TO EXCUTE IDLE TASK", delay)
								// time.Sleep(delay)

								dataSchedule := models.ScheduledMessage{
									To:       *waSession.Contact.Phone,
									Files:    action.Files,
									Message:  parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string)),
									Duration: delay,
									Data: models.WhatsappMessageModel{
										JID:     waSession.JID,
										Message: parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string)),
									},
									Action: &action,
									Task:   &task,
								}

								callback(dataSchedule)

							} else {
								msgData := models.WhatsappMessageModel{
									JID:     waSession.JID,
									Message: parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string)),
								}

								s.customerRelationshipService.WhatsappService.SetMsgData(s.whatsmeowService, &msgData, *waSession.Contact.Phone, action.Files, []models.ProductModel{}, false, nil)
								_, err := customer_relationship.SendCustomerServiceMessage(s.customerRelationshipService.WhatsappService)
								if err != nil {
									log.Println("ERROR", err)
									continue
								}
								task.LastActionTriggerAt = &now
								task.UpdatedAt = &now
								s.ctx.DB.Omit(clause.Associations).Save(&task)
							}

							// thumbnail, restFiles := models.GetThumbnail(action.Files)
							// var fileType, fileUrl string
							// if thumbnail != nil {
							// 	fileType = "image"
							// 	fileUrl = thumbnail.URL
							// }
							// waData := whatsmeow_client.WaMessage{
							// 	JID:      waSession.JID,
							// 	Text:     parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string)),
							// 	To:       *waSession.Contact.Phone,
							// 	IsGroup:  false,
							// 	FileType: fileType,
							// 	FileUrl:  fileUrl,
							// }

							// _, err = s.whatsmeowService.SendMessage(waData)
							// if err != nil {
							// 	continue
							// }

							// for _, v := range restFiles {
							// 	if strings.Contains(v.MimeType, "image") && v.URL != "" {
							// 		resp, _ := s.whatsmeowService.SendMessage(whatsmeow_client.WaMessage{
							// 			JID:      waSession.JID,
							// 			Text:     "",
							// 			To:       *waSession.Contact.Phone,
							// 			IsGroup:  false,
							// 			FileType: "image",
							// 			FileUrl:  v.URL,
							// 		})
							// 		fmt.Println("RESPONSE", resp)
							// 	} else {
							// 		resp, _ := s.whatsmeowService.SendMessage(whatsmeow_client.WaMessage{
							// 			JID:      waSession.JID,
							// 			Text:     "",
							// 			To:       *waSession.Contact.Phone,
							// 			IsGroup:  false,
							// 			FileType: "document",
							// 			FileUrl:  v.URL,
							// 		})
							// 		fmt.Println("RESPONSE", resp)
							// 	}

							// }

							action.ActionStatus = "READY"
							s.ctx.DB.Omit(clause.Associations).Save(&action)

							// msg := parseMsgTemplate(*waSession.Contact, task.CreatedBy, actionData["message"].(string))
							// _, err := sendWAMessage(s.ctx, waSession.JID, *waSession.Contact.Phone, msg)
							// if err != nil {
							// 	fmt.Println("ERROR SENDING MESSAGE", err)
							// 	continue
							// }
							// task.LastActionTriggerAt = &now
							// task.UpdatedAt = &now
							// s.ctx.DB.Omit(clause.Associations).Save(&task)

							// for _, v := range action.Files {
							// 	if strings.Contains(v.MimeType, "image") && v.URL != "" {
							// 		sendWAFileMessage(s.ctx, waSession.JID, *waSession.Contact.Phone, "", "image", v.URL)
							// 	} else {
							// 		sendWAFileMessage(s.ctx, waSession.JID, *waSession.Contact.Phone, "", "document", v.URL)
							// 	}
							// }

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

func sendWAFileMessage(erpContext *context.ERPContext, jid, to, message, fileType, fileUrl string) (any, error) {
	replyData := whatsmeow_client.WaMessage{
		JID:      jid,
		Text:     message,
		To:       to,
		IsGroup:  false,
		FileType: fileType,
		FileUrl:  fileUrl,
	}
	// utils.LogJson(replyData)
	return erpContext.ThirdPartyServices["WA"].(*whatsmeow_client.WhatsmeowService).SendMessage(replyData)
}
