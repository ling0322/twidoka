package main

import (
  "net/http"
  "net/url"
  "github.com/ChimeraCoder/anaconda"
  "fmt"
  // "time"
  "github.com/garyburd/go-oauth/oauth"
  "html/template"
  "strconv"
  // "log"
)

const (
  OK = 0
  RetrieveCount = 200
)

type timelineView struct {
  Title string
  ScreenName string
  PreviousPage int
  NextPage int
  SinceId int64
  Tweets []*tweetView
}

type tweetView struct {
  ScreenName string
  Name string
  InReplyToStatusId int64
  Id int64
  ProfileImageUrl string
  Text template.HTML
  ShowRemove bool
  ShowOperator bool
  CreateTime string
  Source template.HTML
}

type detailsView struct {
  InReplyTo *tweetView
  Tweet *tweetView
  ImageUrl string
  ScreenName string
}

type composeView struct {
  InReplyToTweet *tweetView
  ScreenName string
  DefaultText string
  Type string
}

var homeTemplate = template.Must(template.ParseFiles(
    "templates/madoka.tmpl",
    "templates/menu.tmpl",
    "templates/tweet_list.tmpl",
    "templates/head.tmpl",
    "templates/tweet.tmpl"))

var detailsTemplate = template.Must(template.ParseFiles(
    "templates/details.tmpl",
    "templates/menu.tmpl",
    "templates/head.tmpl",
    "templates/tweet.tmpl"))

var composeTemplate = template.Must(template.ParseFiles(
    "templates/compose.tmpl",
    "templates/menu.tmpl",
    "templates/head.tmpl",
    "templates/tweet.tmpl"))

var errorTemplate = template.Must(template.ParseFiles(
    "templates/error.tmpl",
    "templates/head.tmpl"))

func errorHandler(w http.ResponseWriter, err error) {
  w.WriteHeader(http.StatusInternalServerError)
  errorTemplate.Execute(w, err)
}

func main() {
  http.HandleFunc("/home", homeHandler)
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  page, _ := strconv.Atoi(r.FormValue("page"))
  api := buildAnacondaApiFromRequest(r)

  fmt.Println(r.Host)

  values := buildTimelineRequestValues(r)

  // Gets the timeline from twitter.com
  timeline, err := api.GetHomeTimeline(values)
  tweetViews := make([]*tweetView, len(timeline))
  for i := range timeline {
    tweetViews[i] = convertToTweetView(&timeline[i], screenName, true)
  }

  var lastId int64
  if len(timeline) > 0 {
    lastId = timeline[len(timeline) - 1].Id
  }

  if err == nil {
    err = homeTemplate.Execute(w, &timelineView{
        Title: "Home",
        ScreenName: screenName,
        PreviousPage: page - 1,
        NextPage: page + 1,
        SinceId: lastId,
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
