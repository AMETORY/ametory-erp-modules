package models

type WhatsappContact struct {
	Addresses []WhatsappContactAddress `json:"addresses,omitempty"`
	Birthday  *string                  `json:"birthday,omitempty"`
	Emails    []WhatsappContactEmail   `json:"emails,omitempty"`
	Name      WhatsappContactName      `json:"name,omitempty"`
	Org       *WhatsappContactOrg      `json:"org,omitempty"`
	Phones    []WhatsappContactPhone   `json:"phones,omitempty"`
	Urls      []WhatsappContactUrl     `json:"urls,omitempty"`
}

type WhatsappContactAddress struct {
	Street      *string `json:"street,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	Zip         *string `json:"zip,omitempty"`
	Country     *string `json:"country,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
	Type        *string `json:"type,omitempty"`
}

type WhatsappContactEmail struct {
	Email *string `json:"email,omitempty"`
	Type  *string `json:"type,omitempty"`
}

type WhatsappContactName struct {
	FormattedName *string `json:"formatted_name,omitempty"`
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
	MiddleName    *string `json:"middle_name,omitempty"`
	Suffix        *string `json:"suffix,omitempty"`
	Prefix        *string `json:"prefix,omitempty"`
}

type WhatsappContactOrg struct {
	Company    *string `json:"company,omitempty"`
	Department *string `json:"department,omitempty"`
	Title      *string `json:"title,omitempty"`
}

type WhatsappContactPhone struct {
	Phone *string `json:"phone,omitempty"`
	Type  *string `json:"type,omitempty"`
	WAID  *string `json:"wa_id,omitempty"`
}

type WhatsappContactUrl struct {
	URL  *string `json:"url,omitempty"`
	Type *string `json:"type,omitempty"`
}
