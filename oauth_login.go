package main

import (
  "net/http"
  "net/url"
  "io/ioutil"
  "errors"
  "regexp"
  // "fmt"
)

// Get authenticity_token from Twitter sign in page
func getAuthenticityToken(authorizeUrl string) (token string, err error) {
  resp, err := http.Get(authorizeUrl)
  if err != nil {
    return
  }
  defer resp.Body.Close()
  body_bytes, err := ioutil.ReadAll(resp.Body)
  body := string(body_bytes)
  re := regexp.MustCompile(`<input name="authenticity_token" type="hidden" value="(.+?)" />`)
  match := re.FindStringSubmatch(body)
  if match != nil {
    token = match[1]
  } else {
    err = errors.New("Get authenticity_token failed")
  }
  return
}

func loginAndGetVerifier(username string,
                         password string,
                         oauthToken string,
                         authenticityToken string) (verifier string, err error) {
  data := url.Values{}
  data.Add("repost_after_login", "https://api.twitter.com/oauth/authorize")
  data.Add("authenticity_token", authenticityToken)
  data.Add("oauth_token", oauthToken)
  data.Add("session[username_or_email]", username)
  data.Add("session[password]", password)

  resp, err := http.PostForm("https://api.twitter.com/oauth/authorize", data)
  if err != nil {
    return
  }
  defer resp.Body.Close()
  body_bytes, err := ioutil.ReadAll(resp.Body)
  body := string(body_bytes)

  re := regexp.MustCompile(`<meta http-equiv="refresh" content="0;url=http://(?:.*?)/oauth_token\?oauth_token=(?:.*?)&oauth_verifier=(.*?)">`)
  match := re.FindStringSubmatch(body)
  if match != nil {
    verifier = match[1]
  } else {
    err = errors.New("Login failed")
  }
  return
}