package mgin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

// LoginRsp login response
type LoginRsp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	UserExp      string `json:"user_exp"`
}

type LoginRspExp struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nick_name"`
	LoginIp  string `json:"login_ip"`
}

// LoginRspExpToBase64 login response exp to base64
func LoginRspExpToBase64(exp *LoginRspExp) (string, error) {
	jsonBytes, err := json.Marshal(exp)
	if err != nil {
		return "", err
	}
	base64Encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return base64Encoded, nil
}

// Base64ToLoginRspExp base64 to login response exp
func Base64ToLoginRspExp(base64Str string) (*LoginRspExp, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, errors.New("failed to decode base64 string")
	}

	var loginRspExp LoginRspExp
	if err := json.Unmarshal(decodedBytes, &loginRspExp); err != nil {
		return nil, errors.New("failed to unmarshal JSON bytes into LoginRspExp")
	}

	return &loginRspExp, nil
}
