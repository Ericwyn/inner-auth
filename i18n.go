package main

import (
	"strings"
)

type I18n struct {
	LabelUsername     string
	LabelPassword     string
	LabelTOTP         string
	PlaceholderUsername string
	PlaceholderPassword string
	PlaceholderTOTP   string
	BtnLogin          string
	BtnLogout         string
	ErrInvalid        string
	ErrTOTPRequired   string
	ErrRateLimit      string
	ErrSystemBusy     string
	ErrInternal       string
}

var i18nMap = map[string]I18n{
	"zh": {
		LabelUsername:        "用户名",
		LabelPassword:        "密码",
		LabelTOTP:            "验证码",
		PlaceholderUsername:  "请输入用户名",
		PlaceholderPassword:  "请输入密码",
		PlaceholderTOTP:      "请输入 6 位验证码",
		BtnLogin:             "登录",
		BtnLogout:            "退出登录",
		ErrInvalid:           "用户名或密码错误",
		ErrTOTPRequired:      "请输入验证码",
		ErrRateLimit:         "尝试次数过多，请稍后再试",
		ErrSystemBusy:        "系统繁忙，请稍后再试",
		ErrInternal:          "服务器内部错误",
	},
	"en": {
		LabelUsername:        "Username",
		LabelPassword:        "Password",
		LabelTOTP:            "TOTP Code",
		PlaceholderUsername:  "Enter your username",
		PlaceholderPassword:  "Enter your password",
		PlaceholderTOTP:      "Enter 6-digit code",
		BtnLogin:             "Login",
		BtnLogout:            "Logout",
		ErrInvalid:           "Invalid username or password",
		ErrTOTPRequired:      "TOTP code is required",
		ErrRateLimit:         "Too many attempts, please try again later",
		ErrSystemBusy:        "System busy, please try again later",
		ErrInternal:          "Internal server error",
	},
}

func detectLanguage(acceptLang string) string {
	if strings.Contains(acceptLang, "zh") {
		return "zh"
	}
	return "en"
}

func GetI18n(lang string) I18n {
	if v, ok := i18nMap[lang]; ok {
		return v
	}
	return i18nMap["en"]
}
