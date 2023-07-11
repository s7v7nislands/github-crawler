package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/s7v7nislands/github-crawler/handler"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Port         int
}

func initConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	config := Config{
		ClientID:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		Port:         viper.GetInt("PORT"),
	}

	if config.ClientID == "" {
		return nil, errors.New("client_id is empty")
	}

	if config.ClientSecret == "" {
		return nil, errors.New("client_secret is empty")
	}

	if config.Port == 0 {
		return nil, errors.New("port is empty")
	}

	return &config, nil
}

func main() {
	config, err := initConfig()
	if err != nil {
		fmt.Printf("Init config error: %v\n", err)
		return
	}

	s, err := handler.New(&oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{"user:email", "user:follow"},
		Endpoint:     github.Endpoint,
	})
	if err != nil {
		fmt.Printf("Init handler error: %v\n", err)
		return
	}
	http.HandleFunc("/", s.HandleMain)
	http.HandleFunc("/login", s.HandleGitHubLogin)
	http.HandleFunc("/oauth/callback", s.HandleGitHubCallback)
	http.HandleFunc("/list", s.HandleList)

	// todo: https://prometheus.io/docs/guides/go-application/
	http.Handle("/metrics", promhttp.Handler())
	fmt.Printf("Started running on http://127.0.0.1:%d\n", config.Port)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
