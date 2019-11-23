package services

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/BNPrashanth/poc-go-oauth2/internal/logger"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	oauthConfFb = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:9090/callback-fb",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateStringFb = ""
)

/*
InitializeOAuthFacebook Function
*/
func InitializeOAuthFacebook() {
	oauthConfFb.ClientID = viper.GetString("facebook.clientID")
	oauthConfFb.ClientSecret = viper.GetString("facebook.clientSecret")
	oauthStateStringFb = viper.GetString("oauthStateString")
}

/*
HandleFacebookLogin Function
*/
func HandleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	HandleLogin(w, r, oauthConfFb, oauthStateStringFb)
}

/*
CallBackFromFacebook Function
*/
func CallBackFromFacebook(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Callback-fb..")

	state := r.FormValue("state")
	logger.Log.Info(state)
	if state != oauthStateStringFb {
		logger.Log.Info("invalid oauth state, expected " + oauthStateStringFb + ", got " + state + "\n")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	logger.Log.Info(code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken..\n"))
		reason := r.FormValue("error_reason")
		if reason == "user_denied" {
			w.Write([]byte("User has denied Permission.."))
		}
		// User has denied access..
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		token, err := oauthConfFb.Exchange(oauth2.NoContext, code)
		if err != nil {
			logger.Log.Error("oauthConfFb.Exchange() failed with " + err.Error() + "\n")
			return
		}
		logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
		logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
		logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

		logger.Log.Info("https://graph.facebook.com/me?access_token=" + url.QueryEscape(token.AccessToken) + "&fields=email")
		resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
			url.QueryEscape(token.AccessToken) + "&fields=email")
		if err != nil {
			logger.Log.Error("Get: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Log.Error("ReadAll: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		logger.Log.Info("parseResponseBody: " + string(response) + "\n")

		w.Write([]byte("Hello, I'm protected\n"))
		w.Write([]byte(string(response)))
		return
	}
}
