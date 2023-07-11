package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/s7v7nislands/github-crawler/handler"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Port         int
	RedisAddr    string
}

func initConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := Config{
		ClientID:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		Port:         viper.GetInt("PORT"),
		RedisAddr:    viper.GetString("REDIS_ADDRESS"),
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

	r := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	s, err := handler.New(&oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{"user:email", "user:follow"},
		Endpoint:     github.Endpoint,
	}, r)
	if err != nil {
		fmt.Printf("Init handler error: %v\n", err)
		return
	}
	http.HandleFunc("/", s.HandleMain)
	http.HandleFunc("/login", s.HandleGitHubLogin)
	http.HandleFunc("/oauth/callback", s.HandleGitHubCallback)
	http.HandleFunc("/list", s.HandleList)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Printf("Started running on http://127.0.0.1:%d\n", config.Port)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
