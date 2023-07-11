package handler

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
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
	oauth oauth2.Config
}

func New(oauth oauth2.Config) *Server {
	return &Server{
		oauth: oauth,
	}
}

func (s *Server) HandleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func (s *Server) HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := s.oauth.AuthCodeURL(oauth.StateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle github callback")
	state := r.FormValue("state")
	if state != oauth.StateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'", oauth.StateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := s.oauth.Exchange(r.Context(), code)
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
		oauthClient := s.oauth.Client(r.Context(), value.(*oauth2.Token))
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