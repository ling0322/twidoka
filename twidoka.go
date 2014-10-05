package main

import (
  "net/http"
  "net/url"
  "github.com/ChimeraCoder/anaconda"
  "fmt"
  // "time"
  "github.com/garyburd/go-oauth/oauth"
  // "html/template"
  "strconv"
  // "log"
)

const (
  OK = 0
  RetrieveCount = 200
)

func errorHandler(w http.ResponseWriter, err error) {
  w.WriteHeader(http.StatusInternalServerError)
  errorTemplate.Execute(w, err)
}

func main() {
  http.HandleFunc("/", partialHandler(timelineHandler, "Home"))
  http.HandleFunc("/mentions", partialHandler(timelineHandler, "Mentions"))
  http.HandleFunc("/user", partialHandler(timelineHandler, "User"))
  http.HandleFunc("/signin", signInHandler)
  http.HandleFunc("/oauth_signin", oauthSignInHandler)
  http.HandleFunc("/update", updateHandler)
  http.HandleFunc("/details", detailsHandler)
  http.HandleFunc("/reply", replyHandler)
  http.HandleFunc("/authorize", authorizeHandler)
  http.Handle("/static/", http.FileServer(http.Dir(".")))
  http.HandleFunc("/oauth_token", oauthTokenHandler)
  anaconda.SetConsumerKey(ConsumerKey)
  anaconda.SetConsumerSecret(ConsumerSecret)

  http.ListenAndServe(":8080", nil)
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
  signInTemplate.Execute(w, nil)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
  api := buildAnacondaApiFromRequest(r)

  text := r.FormValue("text")
  inReplyTo := r.FormValue("in_reply_to")

  values := url.Values{}
  if inReplyTo != "" {
    values.Add("in_reply_to_status_id", inReplyTo)
  }

  _, err := api.PostTweet(text, values)
  if err != nil {
    errorHandler(w, err)
  }
}

func replyHandler(w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  api := buildAnacondaApiFromRequest(r)

  var err error
  id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

  var tweet anaconda.Tweet
  if err == nil {
    tweet, err = api.GetTweet(id, url.Values{})
  }

  var view *tweetView
  var compose *composeView
  if err == nil {
    view = convertToTweetView(&tweet, screenName, true)
    compose = new(composeView)
    compose.InReplyToTweet = view
    compose.ScreenName = screenName
    compose.DefaultText = fmt.Sprintf("@%s: ", view.ScreenName)
    compose.Type = "Reply"
    composeTemplate.Execute(w, compose)
  }

  if err != nil {
    errorHandler(w, err)
  }
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  api := buildAnacondaApiFromRequest(r)

  var err error
  id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

  var tweet anaconda.Tweet
  if err == nil {
    tweet, err = api.GetTweet(id, url.Values{})
  }

  var view *tweetView
  var details *detailsView
  var inReplyToView *tweetView
  if err == nil {
    view = convertToTweetView(&tweet, screenName, true)
    if view.InReplyToStatusId > 0 {
      inReplyToTweet, e := api.GetTweet(view.InReplyToStatusId, url.Values{})
      if e == nil {
        inReplyToView = convertToTweetView(&inReplyToTweet, screenName, true)
      }
    }

    details = new(detailsView)
    details.InReplyTo = inReplyToView
    details.Tweet = view
    details.ScreenName = screenName
  }

  if err == nil {
    err = detailsTemplate.Execute(w, details)
  }

  if err != nil {
    errorHandler(w, err)
  }
}

func timelineHandler(timelineType string, w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  page, _ := strconv.Atoi(r.FormValue("page"))
  api := buildAnacondaApiFromRequest(r)
  values := buildTimelineRequestValues(r)

  // Only in user timeline mode
  userScreenName := r.FormValue("u")

  // Gets the timeline from twitter.com
  var err error
  var timeline []anaconda.Tweet
  switch timelineType {
  case "Home":
    timeline, err = api.GetHomeTimeline(values)
  case "Mentions":
    timeline, err = api.GetMentionsTimeline(values)
  case "User":
    values.Add("screen_name", userScreenName)
    timeline, err = api.GetUserTimeline(values)
  }

  // Only in user timeline mode
  var userInf anaconda.User
  var user *userView
  if err == nil && timelineType == "User" {
    userInf, err = api.GetUsersShow(userScreenName, url.Values{})
  }
  if err == nil && timelineType == "User" {
    user = convertToUserView(&userInf)
  }

  if err == nil {
    tweetViews := make([]*tweetView, len(timeline))
    for i := range timeline {
      tweetViews[i] = convertToTweetView(&timeline[i], screenName, true)
    }

    var lastId int64
    if len(timeline) > 0 {
      lastId = timeline[len(timeline) - 1].Id
    }

    timelineTemplate.Execute(w, &timelineView{
        Title: timelineType,
        ScreenName: screenName,
        PreviousPage: page - 1,
        NextPage: page + 1,
        SinceId: lastId,
        User: user,
        Tweets: tweetViews})
  }

  if err != nil {
    errorHandler(w, err)
  }
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
  authUrl, credentials, err := anaconda.AuthorizationURL(SiteURL + "/oauth_token")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    setCookie(w, "access_token", credentials.Token)
    setCookie(w, "access_token_secret", credentials.Secret)
    http.Redirect(w, r, authUrl, http.StatusFound)
  }
}

func oauthSignInHandler(w http.ResponseWriter, r *http.Request) {
  username := r.FormValue("user")
  password := r.FormValue("passwd")

  var err error
  authUrl, credentials, err := anaconda.AuthorizationURL(SiteURL + "/oauth_token")

  var authenticityToken string
  if err == nil {
    setCookie(w, "access_token", credentials.Token)
    setCookie(w, "access_token_secret", credentials.Secret)
    authenticityToken, err = getAuthenticityToken(authUrl)
  }

  var verifier string
  if err == nil {
    verifier, err = loginAndGetVerifier(username, password, credentials.Token, authenticityToken)
  }

  if err == nil {
    oauthTokenUrl := fmt.Sprintf("%s/oauth_token?oauth_verifier=%s", SiteURL, verifier)
    http.Redirect(w, r, oauthTokenUrl, http.StatusFound)
  } else {
    errorHandler(w, err)
  }
}

func oauthTokenHandler(w http.ResponseWriter, r *http.Request) {
  token := getCookie(r, "access_token")
  secret := getCookie(r, "access_token_secret")
  tempCred := &oauth.Credentials{token, secret}

  verifier := r.FormValue("oauth_verifier")
  credentials, values, err := anaconda.GetCredentials(tempCred, verifier)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    setCookie(w, "access_token", credentials.Token)
    setCookie(w, "access_token_secret", credentials.Secret)
    setCookie(w, "screen_name", values["screen_name"][0])
    http.Redirect(w, r, "/", http.StatusFound)
  }
}
