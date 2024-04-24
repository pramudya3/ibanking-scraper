package utils

import (
	"bytes"
	"encoding/json"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const captchaReslvUrl = "https://backend.onpay.co.id/api/v1/captcha"

func ResolveCaptcha() (error, domain.CaptchaResponse) {
	client := &http.Client{}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("captcha", "captcha.png")
	if err != nil {
		logger.Debug("could not create form file:", err)
	}
	file, err := os.Open("captcha.png")
	if err != nil {
		logger.Debug("could not open file:", err)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		logger.Debug("could not copy file:", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", captchaReslvUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		logger.Debug("could not make request post: ", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		logger.Debug("could not make a request: ", err)
	}
	defer res.Body.Close()

	// Tampilkan status dari server
	//fmt.Println(res.Status)
	//logger.Debug(res)
	//logger.Debug(res.Status)
	//Buat variable json response
	var responseJson domain.CaptchaResponse
	// Decode response dari server ke dalam struct
	err = json.NewDecoder(res.Body).Decode(&responseJson)
	if err != nil {
		logger.Debug("err : ", err)
		logger.Debug("could not decode json: ", err)
	}
	return err, responseJson
}
