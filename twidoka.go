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
  "net/http/httputil"
  "strings"
)

const (
  OK = 0
  RetrieveCount = 200
)

func main() {
  http.HandleFunc("/", root(signinRequired(partial(timelineHandler, "Home"))))
  http.HandleFunc("/mentions", signinRequired(partial(timelineHandler, "Mentions")))
  http.HandleFunc("/user", signinRequired(partial(timelineHandler, "User")))
  http.HandleFunc("/search", signinRequired(partial(timelineHandler, "Search")))
  http.HandleFunc("/signin", signInHandler)
  http.HandleFunc("/signout", signOutHandler)
  http.HandleFunc("/update", signinRequired(updateHandler))
  http.HandleFunc("/ajaxupdate", signinRequired(ajaxUpdateHandler))
  http.HandleFunc("/details", signinRequired(detailsHandler))
  http.HandleFunc("/reply", signinRequired(partial(composeHandler, "Reply")))
  http.HandleFunc("/retweet", signinRequired(partial(composeHandler, "Retweet")))
  http.HandleFunc("/compose", signinRequired(partial(composeHandler, "Compose")))
  http.HandleFunc("/remove", signinRequired(removeHandler))
  http.HandleFunc("/authorize", authorizeHandler)
  http.Handle("/static/", http.FileServer(http.Dir(".")))
  http.HandleFunc("/oauth_token", oauthTokenHandler)

  // Add a reverse proxy for images
  dstUrl, _ := url.Parse("http://pbs.twimg.com/")
  reversedProxy := httputil.NewSingleHostReverseProxy(dstUrl)
  http.HandleFunc("/p/", func(w http.ResponseWriter, r *http.Request) {
    r.Host = "pbs.twimg.com"
    // Remove prefix `/p`
    r.URL.Path = r.URL.Path[2: ]
    reversedProxy.ServeHTTP(w, r)
  })

  anaconda.SetConsumerKey(ConsumerKey)
  anaconda.SetConsumerSecret(ConsumerSecret)

  http.ListenAndServe(":80", nil)
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
  deleteCookie(w, "access_token")
  deleteCookie(w, "access_token_secret")
  deleteCookie(w, "screen_name")
  http.Redirect(w, r, "/", http.StatusFound)
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
  var err error
  id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
  confirm := r.FormValue("confirm")
  screenName := getCookie(r, "screen_name")
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()

  if confirm == "" {
    // Not yet confirmed
    var tweet anaconda.Tweet
    if err == nil {
      tweet, err = api.GetTweet(id, url.Values{})
    }
    if err == nil {
      view := convertToTweetView(&tweet, screenName, true)
      removeTemplate.Execute(w, &removeView{
          Tweet: view,
          Referer: r.Referer()})
    }
  } else {
    // Confirmed, remove it!
    if err == nil {
      _, err = api.DeleteTweet(id, true)
    }
    if err == nil {
      referer := r.FormValue("referer")
      if referer == "" {
        referer = "/"
      }
      http.Redirect(w, r, referer, http.StatusFound)
    }
  }

  if err != nil {
    errorHandler(w, err)
  }
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
  signInTemplate.Execute(w, nil)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()

  text := r.FormValue("text")
  inReplyTo := r.FormValue("in_reply_to")

  values := url.Values{}
  if inReplyTo != "" {
    values.Add("in_reply_to_status_id", inReplyTo)
  }

  _, err := api.PostTweet(text, values)
  if err == nil {
    referer := r.FormValue("referer")
    if referer == "" {
      referer = "/"
    }
    http.Redirect(w, r, referer, http.StatusFound)
  } else {
    errorHandler(w, err)
  }
}

func ajaxUpdateHandler(w http.ResponseWriter, r *http.Request) {
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()

  text := r.FormValue("text")
  inReplyTo := r.FormValue("in_reply_to")

  values := url.Values{}
  if inReplyTo != "" {
    values.Add("in_reply_to_status_id", inReplyTo)
  }

  _, err := api.PostTweet(text, values)
  if err == nil {
    fmt.Fprintf(w, "OK")
  } else {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "error: %s", err)
  }
}

func composeHandler(composeType string, w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()

  var err error
  var id int64
  var tweet anaconda.Tweet
  if composeType != "Compose" {
    id, err = strconv.ParseInt(r.FormValue("id"), 10, 64)
    if err == nil {
      tweet, err = api.GetTweet(id, url.Values{})
    }
  }

  if err == nil {
    view := convertToTweetView(&tweet, screenName, true)
    compose := new(composeView)
    compose.ScreenName = screenName
    compose.Referer = r.Referer()
    switch composeType {
    case "Compose":
      compose.DefaultText = ""
    case "Reply":
      replyTo := []string{"@" + view.ScreenName}
      for _, mentionedUser := range view.Mentioned {
        if mentionedUser != screenName {
          replyTo = append(replyTo, "@" + mentionedUser)
        }
      }
      compose.DefaultText = fmt.Sprintf("%s ", strings.Join(replyTo, " "))
      compose.InReplyToTweet = view
    case "Retweet":
      compose.DefaultText = fmt.Sprintf(" RT @%s: %s", view.ScreenName, tweet.Text)
    }
    compose.Type = composeType
    composeTemplate.Execute(w, compose)
  }

  if err != nil {
    errorHandler(w, err)
  }
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
  screenName := getCookie(r, "screen_name")
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()

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
    view.ShowFull = true

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
  api := buildAnacondaApiFromRequest(r)
  defer api.Close()
  values := buildTimelineRequestValues(r)

  // Only in user timeline mode
  userScreenName := r.FormValue("u")

  // Only in search mode
  query := r.FormValue("q")

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
  case "Search":
    if query == "" {
      timeline = make([]anaconda.Tweet, 0)
    } else {
      var searchResponse anaconda.SearchResponse
      searchResponse, err = api.GetSearch(query, values)
      timeline = searchResponse.Statuses
    }
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
        SinceId: lastId,
        User: user,
        Tweets: tweetViews,
        Search: query})
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
    deleteCookie(w, "access_token")
    deleteCookie(w, "access_token_secret")

    url := fmt.Sprintf(
        "?access_token=%s&access_token_secret=%s&screen_name=%s",
        credentials.Token,
        credentials.Secret,
        values["screen_name"][0])
    http.Redirect(w, r, url, http.StatusFound)
  }
}
