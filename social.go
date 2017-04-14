package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	redistore "gopkg.in/boj/redistore.v1"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
)

var oauth2conf = map[string]*oauth2.Config{
	"github": &oauth2.Config{
		ClientID:     appConf.Oauth2.Github.ClientID,
		ClientSecret: appConf.Oauth2.Github.ClientSecret,
		RedirectURL:  prepareRedirectURL(appConf, "github"),
		Endpoint:     github.Endpoint,
	},
	"facebook": &oauth2.Config{
		ClientID:     appConf.Oauth2.Facebook.ClientID,
		ClientSecret: appConf.Oauth2.Facebook.ClientSecret,
		RedirectURL:  prepareRedirectURL(appConf, "facebook"),
		Endpoint:     facebook.Endpoint,
	},
}

func prepareRedirectURL(appConf *conf, name string) string {
	return fmt.Sprintf("http://%s/login/%s/callback", appConf.ProxyHost, name)
}

func serveSocialLogin(store *redistore.RediStore, w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["social"]

	conf, ok := oauth2conf[name]
	log.Println(conf.RedirectURL)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, 301)
}

func serveSocialLoginCallback(store *redistore.RediStore, w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["social"]

	conf, ok := oauth2conf[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := context.Background()

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Errorln("SOCIAL: Empty code.")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Errorln("SOCIAL: Token retrieve error: ", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	js, err := json.MarshalIndent(tok, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "bbg-auth")
	session.Values["social-json"] = js
	session.Save(r, w)

	log.Debugf("SOCIAL: Response - \n %s \n", js)

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
