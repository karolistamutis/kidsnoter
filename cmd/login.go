package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/karolistamutis/kidsnoter/config"
	"github.com/karolistamutis/kidsnoter/logger"
	"github.com/spf13/cobra"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var (
	client     *http.Client
	isLoggedOn bool
)

type loginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func loginPreRun(cmd *cobra.Command, args []string) error {
	if isLoggedOn {
		logger.Log.Debug("Already logged in")
		return nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client = &http.Client{Jar: jar}

	username := config.GetUsername()
	password := config.GetPassword()

	if username == "" || password == "" {
		return fmt.Errorf("username or password is not set in configuration")
	}

	loginData := loginRequest{
		Username:   username,
		Password:   password,
		RememberMe: true,
	}

	jsonLoginData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("failed to marshal login data: %w", err)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.GetAPILoginURL(), bytes.NewBuffer(jsonLoginData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logger.Log.Debugw("Sending login request",
		"url", config.GetAPILoginURL(),
		"username", username,
	)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	var p fastjson.Parser
	v, err := p.Parse(string(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to parse login JSON response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		errCode := string(v.GetStringBytes("err_code"))
		logger.Log.Errorw("Login failed",
			"statusCode", resp.StatusCode,
			"errCode", errCode,
		)
		return fmt.Errorf("login failed: %s", errCode)
	}

	userCookie := &http.Cookie{
		Name:  "current_user",
		Value: username,
	}
	sessionIDCookie := &http.Cookie{
		Name:  "session_id",
		Value: string(v.GetStringBytes("session_id")),
	}

	userCookieDomain, _ := url.Parse(config.GetUserCookieDomain())
	sessionCookieDomain, _ := url.Parse(config.GetSessionCookieDomain())
	client.Jar.SetCookies(userCookieDomain, []*http.Cookie{userCookie})
	client.Jar.SetCookies(sessionCookieDomain, []*http.Cookie{sessionIDCookie})

	isLoggedOn = true

	logger.Log.Info("Successfully logged in")

	return nil
}

// IsLoggedOn returns the login status
func IsLoggedOn() bool {
	return isLoggedOn
}
