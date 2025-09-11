package whatsapp

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"

	"gorm.io/gorm"
)

type msgData struct {
	client    objects.ChatMessage
	data      *models.WhatsappMessageModel
	refMsg    *models.WhatsappMessageModel
	recipient string
	files     []models.FileModel
	products  []models.ProductModel
	saveMsg   bool
}

type WhatsappService struct {
	db      *gorm.DB
	ctx     *context.ERPContext
	msgData *msgData
}

// SetMsgData sets the message data for the whatsapp service.
//
// It takes a client objects.ChatMessage, a message data models.WhatsappMessageModel,
// a recipient string, a slice of files models.FileModel, a slice of products models.ProductModel,
// a boolean to save the message, and an optional reference message models.WhatsappMessageModel.
//
// It will set the message data for the whatsapp service and will be used when sending a message.
func (ws *WhatsappService) SetMsgData(client objects.ChatMessage, data *models.WhatsappMessageModel, recipient string, files []models.FileModel, products []models.ProductModel, saveMsg bool, refMsg *models.WhatsappMessageModel) {
	ws.msgData = &msgData{
		client:    client,
		data:      data,
		recipient: recipient,
		files:     files,
		products:  products,
		saveMsg:   saveMsg,
		refMsg:    refMsg,
	}
}

// NewWhatsappService creates a new instance of WhatsappService.
//
// It initializes the service with a Gorm database and an ERP context,
// allowing for database operations and context-aware processing within the
// service. This setup is essential for managing WhatsApp-related interactions
// and functionalities in the application.

func NewWhatsappService(db *gorm.DB, ctx *context.ERPContext) *WhatsappService {
	return &WhatsappService{
		db:  db,
		ctx: ctx,
	}
}

// CreateWhatsappMessage creates a new WhatsApp message in the database.
//
// This function takes a pointer to a WhatsappMessageModel as an argument
// and inserts it into the database. It returns an error if the operation
// fails, allowing the caller to handle any database-related issues that
// might occur during the creation process.

func (ws *WhatsappService) CreateWhatsappMessage(whatsappMessage *models.WhatsappMessageModel) error {
	return ws.db.Create(whatsappMessage).Error
}

// AddMessageReaction adds or updates a reaction to a WhatsApp message.
//
// This function checks if a reaction already exists for the given
// WhatsApp message and contact. If a reaction exists, it updates the
// reaction with the new value. If no reaction exists, it appends the
// new reaction to the message's reactions.
//
// Parameters:
//   - whatsappMessage: The WhatsApp message to which the reaction belongs.
//   - reaction: The reaction to be added or updated.
//
// Returns:
//   - An error if any database operation fails; otherwise, nil.

func (ws *WhatsappService) AddMessageReaction(whatsappMessage models.WhatsappMessageModel, reaction models.WhatsappMessageReaction) error {
	var existReaction models.WhatsappMessageReaction
	err := ws.db.Where("whatsapp_message_id = ? AND contact_id = ?", whatsappMessage.ID, whatsappMessage.ContactID).First(&existReaction).Error
	if err == nil {
		existReaction.Reaction = reaction.Reaction
		return ws.db.Save(&existReaction).Error
	}
	return ws.db.Model(&whatsappMessage).Association("WhatsappMessageReactions").Append(&reaction)
}

// GetWhatsappMessageByMessageID retrieves a WhatsApp message from the database
// by its message ID. It takes a message ID as a string parameter and returns
// a pointer to a WhatsappMessageModel and an error. If the message is not
// found, the function returns nil and an error.
func (ws *WhatsappService) GetWhatsappMessageByMessageID(messageID string) (*models.WhatsappMessageModel, error) {
	var whatsappMessage models.WhatsappMessageModel
	err := ws.db.Where("message_id = ?", messageID).First(&whatsappMessage).Error
	if err != nil {
		return nil, err
	}
	return &whatsappMessage, nil
}

// GetWhatsappMessageReactions retrieves a list of reactions for a WhatsApp message
// from the database. It takes a WhatsApp message ID as a string parameter and
// returns a slice of WhatsappMessageReaction objects and an error. If no
// reactions are found, the function returns an empty slice and an error.
func (ws *WhatsappService) GetWhatsappMessageReactions(whatsappMessageID string) ([]models.WhatsappMessageReaction, error) {
	var reactions []models.WhatsappMessageReaction
	err := ws.db.Where("whatsapp_message_id = ?", whatsappMessageID).Find(&reactions).Error
	return reactions, err
}

// GetWhatsappMessage retrieves a WhatsApp message from the database by its ID.
//
// It takes an ID as a string parameter and returns a pointer to a WhatsappMessageModel
// and an error. If the message is not found, the function returns nil and an error.
func (ws *WhatsappService) GetWhatsappMessage(id string) (*models.WhatsappMessageModel, error) {
	var whatsappMessage models.WhatsappMessageModel
	err := ws.db.Where("id = ?", id).First(&whatsappMessage).Error
	if err != nil {
		return nil, err
	}
	return &whatsappMessage, nil
}

// UpdateWhatsappMessage updates a WhatsApp message in the database.
//
// This function takes a pointer to a WhatsappMessageModel as an argument
// and updates the corresponding message in the database. It returns an error
// if the operation fails, allowing the caller to handle any database-related
// issues that might occur during the update process.
func (ws *WhatsappService) UpdateWhatsappMessage(whatsappMessage *models.WhatsappMessageModel) error {
	return ws.db.Save(whatsappMessage).Error
}

// DeleteWhatsappMessage deletes a WhatsApp message from the database by its ID.
//
// It takes a message ID as a string parameter and returns an error if the
// operation fails, allowing the caller to handle any database-related issues
// that might occur during the deletion process.
func (ws *WhatsappService) DeleteWhatsappMessage(id string) error {
	return ws.db.Delete(&models.WhatsappMessageModel{}, "id = ?", id).Error
}

// GetWhatsappMessages retrieves a paginated list of WhatsApp messages for a specific JID.
//
// It takes an HTTP request, a search string, and a JID as parameters. The
// search string is optional and can be used to filter the results by sender,
// receiver, or message content. The JID is required and is used to filter
// the results by the JID of the messages.
//
// The function returns a paginated page of WhatsappMessageModel objects and
// an error if the operation fails, allowing the caller to handle any
// database-related issues that might occur during the query process.
//
// The query parameters are as follows:
//
//   - search: an optional string to filter the results by sender, receiver,
//     or message content.
//   - session: an optional string to filter the results by the session name.
func (ws *WhatsappService) GetWhatsappMessages(request http.Request, search string, JID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.db.Model(&models.WhatsappMessageModel{})

	if search != "" {
		stmt = stmt.Where("sender ILIKE ? OR receiver ILIKE ? OR message ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if JID == "" {
		return paginate.Page{}, errors.New("jid is required")
	}

	stmt = stmt.Where("j_id = ?", JID)
	stmt = stmt.Order("created_at DESC")

	if request.URL.Query().Get("session") != "" {
		stmt = stmt.Where("session = ?", request.URL.Query().Get("session"))
	}

	// fmt.Println("KESINI nggak")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageModel{})
	return page, nil

}

// GetWhatsappLastMessages retrieves the most recent WhatsApp message for a specific JID and session.
//
// It takes a JID and a session string as parameters and returns the latest
// WhatsappMessageModel object and an error if the operation fails. The function
// queries the database for the most recently created message that matches the
// given JID and session, ordering results by creation date in descending order.

func (ws *WhatsappService) GetWhatsappLastMessages(JID, session string) (models.WhatsappMessageModel, error) {
	var msg models.WhatsappMessageModel
	stmt := ws.db
	stmt = stmt.Order("created_at DESC").Where("j_id = ? and session = ?", JID, session)
	err := stmt.First(&msg).Error
	return msg, err

}

// GetWhatsappLastCustomerMessages retrieves the most recent WhatsApp message from a customer for a specific JID and session.
//
// It takes a JID, a session string, and a pointer to a WhatsappMessageModel object as parameters and returns an error if the
// operation fails, allowing the caller to handle any database-related issues that might occur during the query process.
//
// The function queries the database for the most recently created message that matches the given JID and session, ordering
// results by creation date in descending order.
func (ws *WhatsappService) GetWhatsappLastCustomerMessages(JID, session string, msg *models.WhatsappMessageModel) error {
	stmt := ws.db
	stmt = stmt.Order("created_at DESC").Where("j_id = ? and session = ?", JID, session)
	return stmt.First(&msg).Error

}

// GetMessageSession retrieves a list of WhatsApp messages grouped by session for a specific JID.
//
// It takes a JID as a string parameter and returns a slice of WhatsappMessageModel objects and
// an error if the operation fails. The function queries the database for all messages that match
// the given JID, groups the results by session name, and returns the most recent message for each
// session. The results are ordered by creation date in descending order.
func (ws *WhatsappService) GetMessageSession(JID string) ([]models.WhatsappMessageModel, error) {

	var waGroup []struct {
		Session string `db:"session"`
	}
	err := ws.db.Model(&models.WhatsappMessageModel{}).Select("session").Where("j_id = ?", JID).Group("session").Find(&waGroup).Error
	if err != nil {
		return []models.WhatsappMessageModel{}, err
	}
	var waMsgs []models.WhatsappMessageModel = []models.WhatsappMessageModel{}
	for _, v := range waGroup {
		var waMsg models.WhatsappMessageModel
		ws.db.Where("j_id = ? AND session = ?", JID, v.Session).Order("created_at DESC").First(&waMsg)
		waMsgs = append(waMsgs, waMsg)
	}
	return waMsgs, nil
}

// GetSessionMessageBySessionName retrieves a paginated list of WhatsApp message sessions for a specific session name, search query, and/or tags.
//
// It takes a session name, an optional search query, and an HTTP request as parameters.
// The function filters message sessions by session name and optionally by search query and tags.
// It returns a paginated page of WhatsappMessageSession models and an error if the operation fails.
//
// The function uses request parameters to modify the pagination and filtering behavior.
// The following query parameters are supported:
//
//   - search: a string to search in the contact's name and email.
//   - tag_ids: a comma-separated list of tag IDs to filter the results.
//   - ID-Company: a header to filter the results by company ID. If the header is
//     set to "nil" or "null", only message sessions with a null company ID are returned.
func (ws *WhatsappService) GetSessionMessageBySessionName(sessionName string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.db.Preload("Contact", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "color")
		}).Select("id", "name", "phone")
	}).Select(
		"whatsapp_message_sessions.id",
		"whatsapp_message_sessions.created_at",
		"whatsapp_message_sessions.updated_at",
		"whatsapp_message_sessions.deleted_at",
		"whatsapp_message_sessions.j_id",
		"whatsapp_message_sessions.session",
		"whatsapp_message_sessions.session_name",
		"whatsapp_message_sessions.last_online_at",
		"whatsapp_message_sessions.last_message",
		"whatsapp_message_sessions.company_id",
		"whatsapp_message_sessions.contact_id",
		"whatsapp_message_sessions.ref_id",
		"whatsapp_message_sessions.ref_type",
		"whatsapp_message_sessions.is_human_agent",
		"whatsapp_message_sessions.is_group",
	).Distinct().Model(&models.WhatsappMessageSession{})
	if sessionName != "" {
		stmt = stmt.Where("session = ?", sessionName)
	}

	if request.URL.Query().Get("search") != "" || request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.
			Joins("LEFT JOIN contacts ON contacts.id = whatsapp_message_sessions.contact_id").
			Joins("LEFT JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
			Joins("LEFT JOIN tags ON tags.id = contact_tags.tag_model_id")
	}

	if request.URL.Query().Get("is_unread") != "" || request.URL.Query().Get("is_unreplied") != "" {
		stmt = stmt.
			Joins("LEFT JOIN whatsapp_messages ON whatsapp_messages.session = whatsapp_message_sessions.session")
		if request.URL.Query().Get("is_unread") != "" && request.URL.Query().Get("is_unreplied") != "" {
			fmt.Println("is_unread && is_unreplied")
			stmt = stmt.Where("(whatsapp_messages.is_read = ? AND whatsapp_messages.is_from_me = ?)  OR (whatsapp_messages.is_replied = ? AND whatsapp_messages.is_from_me = ?)", false, false, false, false)

		} else if request.URL.Query().Get("is_unread") != "" {
			fmt.Println("is_unread")
			stmt = stmt.Where("(whatsapp_messages.is_read = ? AND whatsapp_messages.is_from_me = ?) ", false, false)
		} else if request.URL.Query().Get("is_unreplied") != "" {
			fmt.Println("is_unreplied")
			stmt = stmt.Where("whatsapp_messages.is_replied = ? AND whatsapp_messages.is_from_me = ?", false, false)
		}
	}

	if request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.Where("tags.id in (?)", strings.Split(request.URL.Query().Get("tag_ids"), ","))
	}

	if request.Header.Get("ID-Company") != "" {
		if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
			stmt = stmt.Where("whatsapp_message_sessions.company_id is null")
		} else {
			stmt = stmt.Where("whatsapp_message_sessions.company_id = ?", request.Header.Get("ID-Company"))

		}
	}
	if request.URL.Query().Get("connection_session") != "" {
		stmt = stmt.Where("whatsapp_message_sessions.session_name = ?", request.URL.Query().Get("connection_session"))

	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where(`contacts.name ILIKE ? OR contacts.phone ILIKE ? OR contacts.email ILIKE ?`,
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
		)
	}

	if request.URL.Query().Get("type") != "" {
		if request.URL.Query().Get("type") == "group" {
			stmt = stmt.Where("whatsapp_message_sessions.is_group = ?", true)
		} else if request.URL.Query().Get("type") == "personal" {
			stmt = stmt.Where("whatsapp_message_sessions.is_group = ?", false)
		}

	}

	stmt = stmt.Order("last_online_at DESC")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageSession{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.WhatsappMessageSession)
	newItems := make([]models.WhatsappMessageSession, 0)
	for _, item := range *items {
		if item.Contact == nil {
			continue
		}
		profile, err := item.Contact.GetProfilePicture(ws.ctx.DB)
		if err == nil {
			item.Contact.ProfilePicture = profile
		}

		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

// GetMessageSessionChatBySessionName retrieves a paginated list of WhatsApp messages for a specific session and contact.
//
// It takes a session name, an optional contact ID pointer, and an HTTP request as parameters.
// The function filters messages by session ID and optionally by contact ID. It returns a paginated
// page of WhatsappMessage models and an error if the operation fails.
//
// The function uses request parameters to modify the pagination and filtering behavior.
// The following query parameters are supported:
//
//   - search: a string to search in the contact's name and email.
//   - tag_ids: a comma-separated list of tag IDs to filter the results.
//   - ID-Company: a header to filter the results by company ID. If the header is
//     set to "nil" or "null", only message sessions with a null company ID are returned.
func (ws *WhatsappService) GetMessageSessionChatBySessionName(sessionName string, jid string, contact_id *string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.db.Preload("Member.User").Preload("Contact").Preload("WhatsappMessageReactions").Model(&models.WhatsappMessageModel{})
	fmt.Println("KADIEU")
	if sessionName != "" {
		stmt = stmt.Where("session = ?", sessionName)
	}
	if jid != "" {
		stmt = stmt.Where("j_id = ?", jid)
	}
	if contact_id != nil {
		stmt = stmt.Where("contact_id = ?", *contact_id)
	}

	stmt = stmt.Order("created_at DESC")
	stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageModel{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.WhatsappMessageModel)
	newItems := make([]models.WhatsappMessageModel, 0)
	for _, item := range *items {

		profile, err := item.Contact.GetProfilePicture(ws.ctx.DB)
		if err == nil {
			item.Contact.ProfilePicture = profile
		}

		newItems = append(newItems, item)
	}
	page.Items = &newItems

	return page, nil
}

// ReadAllMessages marks all messages in the session as read.
//
// It takes a session ID as argument and returns an error if the operation
// fails.
func (ws *WhatsappService) ReadAllMessages(session string) error {
	stmt := ws.db.Model(&models.WhatsappMessageModel{}).Where("session = ?", session).Update("is_read", true)
	if stmt.Error != nil {
		return stmt.Error
	}
	return nil
}

// MarkMessageAsRead marks a message as read.
//
// It takes a message ID as argument and returns an error if the operation
// fails.
func (ws *WhatsappService) MarkMessageAsRead(messageId string) error {
	stmt := ws.db.Model(&models.WhatsappMessageModel{}).Where("id= ?", messageId).Update("is_read", true)
	if stmt.RowsAffected == 0 {
		return errors.New("message not found")
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	return nil
}

// GetWhatsappMessageTemplates retrieves a paginated list of WhatsApp message templates from the database.
//
// It takes an HTTP request, a search string, and an optional member ID as parameters. The search string
// filters the message templates by title. The function also allows filtering by company ID from the request
// header and by member ID if provided. The function uses pagination to manage the result set.
//
// The function returns a paginated page of WhatsappMessageTemplate models and an error if the operation fails.

func (s *WhatsappService) GetWhatsappMessageTemplates(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Member").Preload("User")
	if search != "" {
		stmt = stmt.Where("name title ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.Where("member_id = ?", *memberID)
	}
	// request.URL.Query().Get("page")
	stmt = stmt.Model(&models.WhatsappMessageTemplate{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageTemplate{})
	page.Page = page.Page + 1
	return page, nil
}

// GetWhatsappMessageTemplate retrieves a single WhatsApp message template by ID from the database.
//
// It takes the ID of the template as parameter and returns a WhatsappMessageTemplate model and an error if the
// operation fails. If the template is not found, the returned error is "template not found".
//
// The function also retrieves the product images for each product in the template's messages and stores them in
// the ProductImages field of the corresponding product.
func (ws *WhatsappService) GetWhatsappMessageTemplate(ID string) (models.WhatsappMessageTemplate, error) {
	var msg models.WhatsappMessageTemplate
	stmt := ws.db.Where("id = ?", ID)
	err := stmt.First(&msg).Error
	if err != nil {
		return msg, errors.New("template not found")
	}

	for i, message := range msg.Messages {
		for j, v := range message.Products {
			var images []models.FileModel
			ws.db.Where("ref_id = ? and ref_type = ?", v.ID, "product").Find(&images)
			v.ProductImages = images
			msg.Messages[i].Products[j] = v
		}

	}
	return msg, err
}

// AddMessage adds a message to a WhatsApp message template.
//
// It takes the ID of the template as a string and a pointer to a MessageTemplate
// model as parameters. The method sets the WhatsappMessageTemplateID field of the
// message to the given ID and creates the message in the database. If the
// message is not created, the method returns an error with the message "message
// not created".
func (ws *WhatsappService) AddMessage(id string, msg *models.MessageTemplate) error {
	msg.WhatsappMessageTemplateID = &id
	stmt := ws.db.Create(msg)
	if stmt.RowsAffected == 0 {
		return errors.New("message not created")
	}
	return nil
}

// DeleteMessage deletes a message from a WhatsApp message template.
//
// It takes the ID of the template and the message ID as arguments. The method
// deletes the message from the database that matches the given template ID and
// message ID. If no message is found, it returns an error with the message "message
// not found". If the deletion operation encounters an error, it returns the error.

func (ws *WhatsappService) DeleteMessage(id string, messageId string) error {
	stmt := ws.db.Delete(&models.MessageTemplate{}, "whatsapp_message_template_id = ? AND id = ?", id, messageId)
	if stmt.RowsAffected == 0 {
		return errors.New("message not found")
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	return nil
}

// CreateWhatsappMessageTemplate creates a WhatsApp message template and its first message.
//
// The method takes a pointer to a WhatsappMessageTemplate model as a parameter and
// creates the template and its first message in the database. If the template is not
// created, the method returns an error with the message "template not created".
//
// The first message is created with an ID set to a random UUID and the Type set to "whatsapp".
// The method sets the WhatsappMessageTemplateID field of the first message to the ID of the
// created template and creates the first message in the database. If the first message is
// not created, the method returns an error with the message returned by the database.
func (ws *WhatsappService) CreateWhatsappMessageTemplate(msg *models.WhatsappMessageTemplate) error {
	var firstMsg models.MessageTemplate
	firstMsg.ID = utils.Uuid()
	firstMsg.Type = "whatsapp"

	stmt := ws.db.Create(msg)
	if stmt.RowsAffected == 0 {
		return errors.New("template not created")
	}
	if stmt.Error != nil {
		return stmt.Error
	}

	firstMsg.WhatsappMessageTemplateID = &msg.ID
	return ws.ctx.DB.Create(&firstMsg).Error
}

// UpdateWhatsappMessageTemplate updates a WhatsApp message template.
//
// It takes the ID of the template and a pointer to a WhatsappMessageTemplate model as parameters.
// The method updates the template in the database that matches the given ID with the given model.
// If no template is found, it returns an error with the message "template not found".
// If the update operation encounters an error, it returns the error.
func (ws *WhatsappService) UpdateWhatsappMessageTemplate(id string, msg *models.WhatsappMessageTemplate) error {
	stmt := ws.db.Model(&models.WhatsappMessageTemplate{}).Where("id = ?", id).Updates(msg)
	if stmt.RowsAffected == 0 {
		return errors.New("template not found")
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	return nil
}

// AddProductWhatsappMessageTemplate adds a product to a WhatsApp message template
//
// It takes the ID of the template, the ID of the message and a pointer to a ProductModel as parameters.
// The method first checks if the template and the message with the given IDs exist in the database.
// If either of them do not exist, the method returns an error with the message "template not found"
// or "msg not found" respectively. If both exist, the method clears all the products associated with
// the message and then appends the given product to the message in the database. If the append
// operation encounters an error, the method returns the error.
func (ws *WhatsappService) AddProductWhatsappMessageTemplate(templateID string, ID string, product *models.ProductModel) error {
	var template models.WhatsappMessageTemplate
	stmt := ws.db.Where("id = ?", templateID).First(&template)
	if stmt.RowsAffected == 0 {
		return errors.New("template not found")
	}

	var msg models.MessageTemplate
	stmt = ws.db.Where("id = ?", ID).First(&msg)
	if stmt.RowsAffected == 0 {
		return errors.New("msg not found")
	}
	ws.db.Model(&msg).Association("Products").Clear()

	return ws.db.Model(&msg).Association("Products").Append(product)
}

// DeleteWhatsappMessageTemplate deletes a WhatsApp message template with the given ID from the database.
//
// If no template with the given ID is found, it returns an error with the message "template not found".
// If the deletion operation encounters an error, it returns the error.
func (ws *WhatsappService) DeleteWhatsappMessageTemplate(ID string) error {
	stmt := ws.db.Delete(&models.WhatsappMessageTemplate{}, "id = ?", ID)
	if stmt.RowsAffected == 0 {
		return errors.New("template not found")
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	return nil
}

// SendCSMessage sends a customer service message using the input provided to the Whatsapp service.
//
// It first checks if the input is nil and returns an error if it is. Then, it calls SendWhatsappMessage
// to send the message and processes the response to extract the message data, which is stored in the input data.
// If the input specifies that the message should be saved, it calls SaveMessage to save the message data.
//
// Returns the message data if successful, or an error if any step fails.
func (ws *WhatsappService) SendCSMessage() (any, error) {
	if ws.msgData == nil {
		return nil, errors.New("msgData is nil")
	}
	err := ws.SendWhatsappMessage(ws.msgData.client, ws.msgData.data, ws.msgData.recipient, ws.msgData.files, ws.msgData.products, ws.msgData.saveMsg)
	if err != nil {
		return nil, err
	}
	return ws.msgData.data, nil
}

// SendWhatsappMessage sends a message to the given recipient using the provided WhatsApp client.
//
// It takes a client objects.ChatMessage, a message data models.WhatsappMessageModel,
// a recipient string, a slice of files models.FileModel, a slice of products models.ProductModel,
// and a boolean to save the message.
//
// It will send the message and process the response to extract the message data,
// which is stored in the input data. If the input specifies that the message should be saved,
// it calls SaveMessage to save the message data.
//
// Returns an error if any step fails.
func (ws *WhatsappService) SendWhatsappMessage(client objects.ChatMessage,
	data *models.WhatsappMessageModel,
	recipient string,
	files []models.FileModel,
	products []models.ProductModel,
	saveMsg bool,
) error {
	if client == nil {
		return errors.New("client is nil")
	}
	thumbnail, restFiles := models.GetThumbnail(files)
	var fileType, fileUrl string

	if thumbnail != nil {
		fileType = "image"
		fileUrl = thumbnail.URL
		data.MediaURL = thumbnail.URL
		data.MimeType = thumbnail.MimeType
	}

	msgData := whatsmeow_client.WaMessage{
		JID:      data.JID,
		Text:     data.Message,
		To:       recipient,
		IsGroup:  data.IsGroup,
		FileType: fileType,
		FileUrl:  fileUrl,
	}

	if ws.msgData.refMsg != nil {
		msgData.RefID = ws.msgData.refMsg.MessageID
		info := ws.msgData.refMsg.MessageInfo
		if info != nil {
			from := info["Chat"].(string)
			msgData.RefFrom = &from
		} else {
			from := ws.msgData.refMsg.Sender + "@s.whatsapp.net"
			msgData.RefFrom = &from
		}
		msgData.RefText = &ws.msgData.refMsg.Message

	}

	msgID := ""
	if newClient, ok := client.(*whatsmeow_client.WhatsmeowService); ok {
		utils.LogJson(msgData)
		newClient.SetChatData(msgData)
		resp, err := objects.SendChatMessage(newClient)
		if err != nil {
			log.Println(err)
			return err
		}
		if resp != nil {
			// msgID = resp.(map[string]any)["data"].(map[string]any)["ID"].(string)
			respData, ok := resp.(map[string]any)["data"].(map[string]any)
			if ok {
				msgID, _ = respData["ID"].(string)
			}
		}

	}

	if saveMsg {
		data.MessageID = &msgID
		err := ws.CreateWhatsappMessage(data)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	for _, v := range restFiles {
		msgFile := whatsmeow_client.WaMessage{
			JID:     data.JID,
			To:      recipient,
			IsGroup: data.IsGroup,
			FileUrl: v.URL,
		}
		if strings.Contains(v.MimeType, "image") && v.URL != "" {
			msgFile.FileType = "image"
		} else {
			msgFile.FileType = "document"
		}
		msgID := ""
		if newClient, ok := client.(*whatsmeow_client.WhatsmeowService); ok {
			newClient.SetChatData(msgFile)
			resp, err := objects.SendChatMessage(newClient)
			if err != nil {
				log.Println(err)
				return err
			}
			respData, ok := resp.(map[string]any)["data"].(map[string]any)
			if ok {
				msgID, _ = respData["ID"].(string)
			}
		}
		msgFileData := *data
		msgFileData.ID = utils.Uuid()
		msgFileData.MediaURL = v.URL
		msgFileData.MimeType = v.MimeType
		msgFileData.Message = ""
		msgFileData.MessageID = &msgID
		if saveMsg {
			err := ws.CreateWhatsappMessage(&msgFileData)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}

	for _, v := range products {
		desc := ""
		var images []models.FileModel
		ws.db.Where("ref_id = ? and ref_type = ?", v.ID, "product").Find(&images)

		if v.Description != nil {
			desc = *v.Description
		}
		dataMsg := fmt.Sprintf(`*%s*
_%s_

%s
				`, v.DisplayName, utils.FormatRupiah(v.Price), desc)
		productMsg := whatsmeow_client.WaMessage{
			JID:      data.JID,
			Text:     dataMsg,
			To:       recipient,
			IsGroup:  data.IsGroup,
			FileType: "image",
		}
		if len(images) > 0 {
			productMsg.FileType = "image"
			productMsg.FileUrl = images[0].URL
		}
		msgID := ""
		if newClient, ok := client.(*whatsmeow_client.WhatsmeowService); ok {
			newClient.SetChatData(productMsg)
			resp, err := objects.SendChatMessage(newClient)
			if err != nil {
				log.Println(err)
				return err
			}
			respData, ok := resp.(map[string]any)["data"].(map[string]any)
			if ok {
				msgID, _ = respData["ID"].(string)
			}
		}

		msgProductData := *data
		msgProductData.ID = utils.Uuid()
		msgProductData.Message = dataMsg
		msgProductData.MessageID = &msgID
		if len(images) > 0 {
			msgProductData.MediaURL = images[0].URL
			msgProductData.MimeType = images[0].MimeType
		}
		if saveMsg {
			err := ws.CreateWhatsappMessage(&msgProductData)
			if err != nil {
				log.Println(err)
				return err
			}
		}

	}

	return nil
}
