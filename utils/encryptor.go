package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/pkg/constant"
	"io"
)

func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	cipherText := make([]byte, aes.BlockSize+len(b))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(b))

	return cipherText, nil
}

func Decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		logger.Debug("could not decode data :", err)
	}
	return data, nil
}

func EncodePassword(password string) string {
	passwordEncrypted, err := Encrypt([]byte(constant.Key), []byte(password))
	if err != nil {
		logger.Debug("could not encrypt password :", err)
	}
	passwordEncoded := base64.StdEncoding.EncodeToString(passwordEncrypted)
	return passwordEncoded
}

func DecodePassword(password string) string {
	passwordDecoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		logger.Error(err)
	}
	passwordDecrypted, err := Decrypt([]byte(constant.Key), passwordDecoded)
	if err != nil {
		logger.Error(err)
	}
	return string(passwordDecrypted)
}
