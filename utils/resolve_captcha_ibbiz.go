package utils

import (
	"bytes"
	"encoding/json"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/pkg/constant"
	"net/http"

	api2captcha "github.com/2captcha/2captcha-go"
)

const (
	urlCaptchaAudio = "https://stt-green.vercel.app/api/stt"
	token           = "NJXJMYPNN4G6G7FYFJ5V3PED6QMSPDMX"
)

type (
	captchaRequest struct {
		Url   string `json:"url"`
		Token string `json:"token"`
	}

	captchaResponse struct {
		Text string `json:"text"`
		Code string `json:"code"`
	}
)

func CaptchaResolver() string {
	client := api2captcha.NewClient(constant.ApiKey)

	cap := api2captcha.ReCaptcha{
		SiteKey:   "6LfQ9OkUAAAAABLHOt-u-7X662tf_dBqR0EeYHbw",
		Url:       "https://newbiz.bri.co.id/",
		Invisible: true,
		Action:    "verify",
	}

	req := cap.ToRequest()
	code, err := client.Solve(req)
	if err != nil {
		logger.Debug("could not get code captcha :", err)
	}

	return code
}

func CaptchaSoundResolver(url string) (error, captchaResponse) {

	req, _ := json.Marshal(map[string]string{
		"url":   url,
		"token": token,
	})
	//reqBody := bytes.NewBuffer(req)

	resp, err := http.Post(urlCaptchaAudio, "application/json", bytes.NewReader(req))
	if err != nil {
		logger.Debug("error an occured :", err)
	}
	defer resp.Body.Close()

	//if resp.Status != "200 OK" {
	//	logger.Debug("Proses Login")
	//}

	var responseJson captchaResponse
	err = json.NewDecoder(resp.Body).Decode(&responseJson)
	if err != nil {
		logger.Debug("error decode json :", err)
	}

	return err, responseJson
}
