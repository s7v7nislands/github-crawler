package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/s7v7nislands/github-crawler/handler"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func main() {
	config := oauth2.Config{
		ClientID:     os.Getenv("ClIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"user:email", "user:follow"},
		Endpoint:     github.Endpoint,
	}
	s := handler.New(config)
	http.HandleFunc("/", s.HandleMain)
	http.HandleFunc("/login", s.HandleGitHubLogin)
	http.HandleFunc("/oauth/callback", s.HandleGitHubCallback)
	http.HandleFunc("/list", s.HandleList)

	// todo: https://prometheus.io/docs/guides/go-application/
	http.Handle("/metrics", promhttp.Handler())
	fmt.Print("Started running on http://127.0.0.1:9090\n")
	fmt.Println(http.ListenAndServe(":9090", nil))
}
