package handler

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/redis/go-redis/v9"
	"github.com/s7v7nislands/github-crawler/metrics"
	"github.com/s7v7nislands/github-crawler/oauth"
	"golang.org/x/oauth2"
)

const htmlIndex = `<html><body>
Logged in with <a href="/login">GitHub</a>
</body></html>
`

const htmlList = `<html><body>
Welcome!
<table>
  <tr>
    <th>username</th>
    <th>followers</th>
  </tr>
  {{ range .}}
  <tr>
    <td>{{ .Name }}</td>
    <td>{{ .Followers }}</td>
  </tr>
  {{ end }}
</table>

</body></html>
`

var tokens sync.Map

type Server struct {
	oauthConfig *oauth2.Config
	stateCache  *oauth.StateCache
}

func New(config *oauth2.Config, r *redis.Client) (*Server, error) {
	cache := oauth.NewStateCache(r, 5*time.Minute)
	return &Server{
		oauthConfig: config,
		stateCache:  cache,
	}, nil
}

func (s *Server) HandleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlIndex))
}

func (s *Server) HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := s.stateCache.SetState(r.Context())
	if err != nil {
		log.Printf("set state error: %s", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	url := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle github callback")
	state := r.FormValue("state")
	x, err := s.stateCache.GetDelState(r.Context(), state)
	if err != nil {
		log.Printf("get state error: %s", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if x != "" {
		log.Printf("invalid oauth state got '%s'", state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := s.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	metrics.OpsProcessed.Inc()

	tokens.Store(token.AccessToken, token)
	http.Redirect(w, r, "/list", http.StatusTemporaryRedirect)
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	infos := []*github.User{}
	tokens.Range(func(key, value any) bool {
		oauthClient := s.oauthConfig.Client(r.Context(), value.(*oauth2.Token))
		client := github.NewClient(oauthClient)
		user, _, err := client.Users.Get(r.Context(), "")
		if err != nil {
			log.Printf("client.Users.Get() faled with '%s'\n", err)
			return false
		}
		log.Printf("Logged in as GitHub user: %s", *user.Login)
		infos = append(infos, user)
		return true
	})
	t := template.Must(template.New("foo").Parse(htmlList))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := t.Execute(w, infos)
	if err != nil {
		log.Printf("Template execute: %s", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
