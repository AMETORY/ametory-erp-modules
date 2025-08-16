package objects

import (
	"time"
)

type WhatsAppMessage struct {
	Conversation        *string          `json:"conversation"`
	ImageMessage        *ImageMessage    `json:"imageMessage,omitempty"`
	VideoMessage        *VideoMessage    `json:"videoMessage,omitempty"`
	DocumentMessage     *DocumentMessage `json:"documentMessage,omitempty"`
	LocationMessage     *LocationMessage `json:"locationMessage,omitempty"`
	LiveLocationMessage *LocationMessage `json:"liveLocationMessage,omitempty"`
	ContactMessage      *ContactMessage  `json:"contactMessage,omitempty"`
	MessageContextInfo  struct {
		DeviceListMetadata struct {
			SenderKeyHash      string `json:"senderKeyHash"`
			SenderTimestamp    int64  `json:"senderTimestamp"`
			RecipientKeyHash   string `json:"recipientKeyHash"`
			RecipientTimestamp int64  `json:"recipientTimestamp"`
		} `json:"deviceListMetadata"`
		DeviceListMetadataVersion int    `json:"deviceListMetadataVersion"`
		MessageSecret             string `json:"messageSecret"`
	} `json:"messageContextInfo"`
	ExtendedTextMessage  *ExtendedTextMessage  `json:"extendedTextMessage,omitempty"`
	ReactionMessage      *ReactionMessage      `json:"reactionMessage,omitempty"`
	EventMessage         *EventMessage         `json:"eventMessage,omitempty"`
	ContactsArrayMessage *ContactsArrayMessage `json:"contactsArrayMessage,omitempty"`
}

type LocationMessage struct {
	DegreesLatitude  float64 `json:"degreesLatitude"`
	DegreesLongitude float64 `json:"degreesLongitude"`
	JPEGThumbnail    string  `json:"JPEGThumbnail"`
	SequenceNumber   int64   `json:"sequenceNumber"`
}

type EventMessage struct {
	IsCanceled         *bool            `json:"isCanceled,omitempty"`
	Name               *string          `json:"name,omitempty"`
	Description        *string          `json:"description,omitempty"`
	Location           *LocationMessage `json:"location,omitempty"`
	JoinLink           *string          `json:"joinLink,omitempty"`
	StartTime          *int64           `json:"startTime,omitempty"`
	EndTime            *int64           `json:"endTime,omitempty"`
	ExtraGuestsAllowed *bool            `json:"extraGuestsAllowed,omitempty"`
	IsScheduleCall     *bool            `json:"isScheduleCall,omitempty"`
}

type ContactsArrayMessage struct {
	DisplayName *string           `json:"displayName,omitempty"`
	Contacts    []*ContactMessage `json:"contacts,omitempty"`
}

type ContactMessage struct {
	DisplayName *string `json:"displayName,omitempty"`
	Vcard       *string `json:"vcard,omitempty"`
}
type ReactionMessage struct {
	Key struct {
		RemoteJID string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		ID        string `json:"id"`
	} `json:"key"`
	Text              string `json:"text"`
	SenderTimestampMS int64  `json:"senderTimestampMS"`
}

type ExtendedTextMessage struct {
	Text        string             `json:"text"`
	ContextInfo MessageContextInfo `json:"contextInfo"`
}

type QuotedMessage struct {
	Conversation string `json:"conversation"`
}

type MessageContextInfo struct {
	StanzaID      string        `json:"stanzaID"`
	Participant   string        `json:"participant"`
	QuotedMessage QuotedMessage `json:"quotedMessage"`
}

type ImageMessage struct {
	URL                  string `json:"URL"`
	Mimetype             string `json:"mimetype"`
	Caption              string `json:"caption"`
	FileSHA256           string `json:"fileSHA256"`
	FileLength           int    `json:"fileLength"`
	Height               int    `json:"height"`
	Width                int    `json:"width"`
	MediaKey             string `json:"mediaKey"`
	FileEncSHA256        string `json:"fileEncSHA256"`
	DirectPath           string `json:"directPath"`
	MediaKeyTimestamp    int64  `json:"mediaKeyTimestamp"`
	JPEGThumbnail        string `json:"JPEGThumbnail"`
	FirstScanSidecar     string `json:"firstScanSidecar"`
	FirstScanLength      int    `json:"firstScanLength"`
	ScansSidecar         string `json:"scansSidecar"`
	ScanLengths          []int  `json:"scanLengths"`
	MidQualityFileSHA256 string `json:"midQualityFileSHA256"`
	ImageSourceType      int    `json:"imageSourceType"`
}

type VideoMessage struct {
	URL                string `json:"URL"`
	Mimetype           string `json:"mimetype"`
	Caption            string `json:"caption"`
	FileSHA256         string `json:"fileSHA256"`
	FileLength         int    `json:"fileLength"`
	Seconds            int    `json:"seconds"`
	MediaKey           string `json:"mediaKey"`
	Height             int    `json:"height"`
	Width              int    `json:"width"`
	FileEncSHA256      string `json:"fileEncSHA256"`
	DirectPath         string `json:"directPath"`
	MediaKeyTimestamp  int64  `json:"mediaKeyTimestamp"`
	JPEGThumbnail      string `json:"JPEGThumbnail"`
	StreamingSidecar   string `json:"streamingSidecar"`
	AccessibilityLabel string `json:"accessibilityLabel"`
}

type DocumentMessage struct {
	URL               string `json:"URL"`
	Mimetype          string `json:"mimetype"`
	Title             string `json:"title"`
	FileSHA256        string `json:"fileSHA256"`
	FileLength        int    `json:"fileLength"`
	MediaKey          string `json:"mediaKey"`
	FileName          string `json:"fileName"`
	FileEncSHA256     string `json:"fileEncSHA256"`
	DirectPath        string `json:"directPath"`
	MediaKeyTimestamp int64  `json:"mediaKeyTimestamp"`
	ContactVcard      bool   `json:"contactVcard"`
	Caption           string `json:"caption"`
}

type MsgObject struct {
	Message     WhatsAppMessage `json:"message"`
	Info        map[string]any  `json:"info"`
	GroupInfo   GroupInfo       `json:"group_info"`
	Sender      string          `json:"sender"`
	JID         string          `json:"jid"`
	SessionID   string          `json:"session_id"`
	SessionName string          `json:"session_name"`
	MediaPath   string          `json:"media_path"`
	MimeType    string          `json:"mime_type"`
	ProfilePic  string          `json:"profile_pic"`
}

type GroupInfo struct {
	JID                           string    `json:"JID"`
	OwnerJID                      string    `json:"OwnerJID"`
	Name                          string    `json:"Name"`
	NameSetAt                     time.Time `json:"NameSetAt"`
	NameSetBy                     string    `json:"NameSetBy"`
	Topic                         string    `json:"Topic"`
	TopicID                       string    `json:"TopicID"`
	TopicSetAt                    time.Time `json:"TopicSetAt"`
	TopicSetBy                    string    `json:"TopicSetBy"`
	TopicDeleted                  bool      `json:"TopicDeleted"`
	IsLocked                      bool      `json:"IsLocked"`
	IsAnnounce                    bool      `json:"IsAnnounce"`
	AnnounceVersionID             string    `json:"AnnounceVersionID"`
	IsEphemeral                   bool      `json:"IsEphemeral"`
	DisappearingTimer             int       `json:"DisappearingTimer"`
	IsIncognito                   bool      `json:"IsIncognito"`
	IsParent                      bool      `json:"IsParent"`
	DefaultMembershipApprovalMode string    `json:"DefaultMembershipApprovalMode"`
	LinkedParentJID               string    `json:"LinkedParentJID"`
	IsDefaultSubGroup             bool      `json:"IsDefaultSubGroup"`
	IsJoinApprovalRequired        bool      `json:"IsJoinApprovalRequired"`
	GroupCreated                  time.Time `json:"GroupCreated"`
	ParticipantVersionID          string    `json:"ParticipantVersionID"`
	Participants                  []struct {
		JID          string      `json:"JID"`
		LID          string      `json:"LID"`
		IsAdmin      bool        `json:"IsAdmin"`
		IsSuperAdmin bool        `json:"IsSuperAdmin"`
		DisplayName  string      `json:"DisplayName"`
		Error        int         `json:"Error"`
		AddRequest   interface{} `json:"AddRequest"`
	} `json:"Participants"`
	MemberAddMode string `json:"MemberAddMode"`
}
