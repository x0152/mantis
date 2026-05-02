package api

import usecases "mantis/apps/telegram/use_cases"

type WizardVerifyInput struct {
	Body struct {
		Token string `json:"token" required:"true" minLength:"1"`
	}
}

type WizardVerifyOutput struct {
	Body usecases.WizardBot
}

type WizardStatusInput struct {
	Body struct {
		Token string `json:"token" required:"true" minLength:"1"`
	}
}

type WizardStatusOutput struct {
	Body struct {
		User *usecases.WizardUser `json:"user"`
	}
}
