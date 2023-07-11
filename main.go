package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/s7v7nislands/github-crawler/handler"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func initConfig() (*oauth2.Config, error) {
	e := os.Environ()
	for _, i := range e {
		fmt.Printf("%s\n", i)
	}
	config := oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"user:email", "user:follow"},
		Endpoint:     github.Endpoint,
	}

	if config.ClientID == "" {
		return nil, errors.New("client_id is empty")
	}

	if config.ClientSecret == "" {
		return nil, errors.New("client_secret is empty")
	}

	return &config, nil
}

func main() {
	config, err := initConfig()
	if err != nil {
		fmt.Printf("Init config error: %v\n", err)
		return
	}

	s, err := handler.New(config)
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
	fmt.Printf("Started running on http://127.0.0.1:9090\n")
	fmt.Println(http.ListenAndServe(":9090", nil))
}
