package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// InitConfig initializes the configuration
func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/config")
	viper.SetEnvPrefix("KIDSNOTER")

	viper.AutomaticEnv()

	// Set some defaults
	const targetURL = "https://www.kidsnote.com/api"
	viper.SetDefault("api.base_url", targetURL)
	viper.SetDefault("api.login_url", targetURL+"/web/login")
	viper.SetDefault("api.info_url", targetURL+"/v1/me/info")
	viper.SetDefault("api.album_url", targetURL+"/v1_2/children/%d/albums")
	viper.SetDefault("cookies.user_domain", "www.kidsnote.com")
	viper.SetDefault("cookies.session_domain", ".kidsnote.com")

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			fmt.Println("No config file found. Using defaults and environment variables.")
		}
	}
	return nil
}

// GetString retrieves a string configuration value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt retrieves an integer configuration value
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool retrieves a boolean configuration value
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetAPIBaseURL returns the base URL for the API
func GetAPIBaseURL() string {
	return viper.GetString("api.base_url")
}

// GetAPILoginURL returns the login URL for the API
func GetAPILoginURL() string {
	return viper.GetString("api.login_url")
}

// GetAPIInfoURL returns the profile info URL for the API
func GetAPIInfoURL() string {
	return viper.GetString("api.info_url")
}

// GetAPIAlbumURL returns the album retrieval URL for the API
func GetAPIAlbumURL() string {
	return viper.GetString("api.album_url")
}

// GetUserCookieDomain returns the domain to be used for storing login related cookie data
func GetUserCookieDomain() string {
	return viper.GetString("cookies.user_domain")
}

// GetSessionCookieDomain returns the domain to be used for storing session_id related cookie data
func GetSessionCookieDomain() string {
	return viper.GetString("cookies.session_domain")
}

// GetUsername returns the logon username of the API
func GetUsername() string {
	return viper.GetString("username")
}

// GetPassword returns the logon password of the API
func GetPassword() string {
	return viper.GetString("password")
}

// GetAlbumDir returns the path to the directory root for saving downloaded albums
func GetAlbumDir() string { return viper.GetString("album_dir") }

// GetSyncInterval returns the album sync interval in hours
func GetSyncInterval() time.Duration { return viper.GetDuration("sync_interval") }
