package bot

import (
	"encoding/json"

	_ "embed"
)

const (
	UALang = "UA"
)

const (
	userRoleRequestTranslation     = "user_role_request"
	userCategoryRequest            = "user_category_request"
	userLocalityRequestTranslation = "user_locality_request"
	userLocalityReplyTranslation   = "user_locality_reply"
	contactPhoneRequestTranslation = "contact_phone_request"

	btnOptionUserRoleSeeker    = "btn_option_user_role_seeker"
	btnOptionUserRoleVolunteer = "btn_option_user_role_volunteer"

	errorChooseOption = "error_choose_option"

	volunteerChosenCategoriesHeaderTr = "volunteer_chosen_categories_header"
	volunteerChosenCategoriesFooterTr = "volunteer_chosen_categories_footer"
	nextButtonTr                      = "next_button"

	helpCategoriesTranslation = "help_categories_reply"
	helpLocalityTranslation   = "help_location_reply"
	helpCreateAtTranslation   = "help_created_at_reply"
	helpDetailsTranslation    = "help_details_translation_reply"
	helpsEmptyTranslation     = "helps_empty_reply"

	subscriptionRequestTranslation = "subscription_request_translation"
	subscriptionButtonTranslation  = "subscription_button_translation"
)

//go:embed translation.json
var translations []byte // nolint:gochecknoglobals

type Translator interface {
	Translate(key, lang string) string
}

type Tr map[string]map[string]string

func (t Tr) Translate(key, lang string) string {
	return t[key][lang]
}

func NewTranslator() (Tr, error) {
	var trmap = make(map[string]map[string]string)
	return trmap, json.Unmarshal(translations, &trmap)
}
