package dto

// CreateAuthorDTO defines the structure for creating a new author.
// @name CreateAuthor
type CreateAuthorDTO struct {
	Name    TranslatedTextDTO `json:"name" binding:"required"`
	Bio     TranslatedTextDTO `json:"bio"`
	Contact struct {
		Email string `json:"email" binding:"email" example:"vovka@sosijopa.com"`
		Phone string `json:"phone" example:"+5553535"`
		Links struct {
			Instagram string `json:"instagram" example:"vovka_insta"`
			Telegram  string `json:"telegram" example:"vovka_tele"`
			Vkontakte string `json:"vkontakte" example:"vovka_vk"`
			Facebook  string `json:"facebook" example:"vovka_fb"`
			Twitter   string `json:"twitter" example:"vovka_tweet"`
			Youtube   string `json:"youtube" example:"vovka_channel"`
			Linkedin  string `json:"linkedin" example:"vovka_linkedin"`
			Whatsapp  string `json:"whatsapp" example:"+5553535"`
			Pinterest string `json:"pinterest" example:"vovka_pins"`
			Behance   string `json:"behance" example:"vovka_art"`
		} `json:"links"`
	} `json:"contact"`
	IsActive bool `json:"is_active" example:"true"`
}

// UpdateAuthorDTO defines the structure for updating an existing author.
// @name UpdateAuthor
type UpdateAuthorDTO struct {
	Name    TranslatedTextDTO `json:"name"`
	Bio     TranslatedTextDTO `json:"bio"`
	Contact struct {
		Email string `json:"email" binding:"email" example:"vovka@sosijopa.com"`
		Phone string `json:"phone" example:"+5553535"`
		Links struct {
			Instagram string `json:"instagram" example:"vovka_insta"`
			Telegram  string `json:"telegram" example:"vovka_tele"`
			Vkontakte string `json:"vkontakte" example:"vovka_vk"`
			Facebook  string `json:"facebook" example:"vovka_fb"`
			Twitter   string `json:"twitter" example:"vovka_tweet"`
			Youtube   string `json:"youtube" example:"vovka_channel"`
			Linkedin  string `json:"linkedin" example:"vovka_linkedin"`
			Whatsapp  string `json:"whatsapp" example:"+5553535"`
			Pinterest string `json:"pinterest" example:"vovka_pins"`
			Behance   string `json:"behance" example:"vovka_art"`
		} `json:"links"`
	} `json:"contact"`
	IsActive bool `json:"is_active" example:"false"`
}
