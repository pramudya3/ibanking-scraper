package domain

type CaptchaResponse struct {
	Data struct {
		Captcha string `json:"captcha"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}
