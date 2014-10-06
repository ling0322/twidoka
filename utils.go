package main

import (
  "net/http"
  "github.com/ChimeraCoder/anaconda"
  "time"
  "fmt"
  "strconv"
  "html/template"
  "net/url"
  "strings"
)

type notFoundError struct {
}
func (e *notFoundError) Error() string {
  return "Page not found"
}

func errorHandler(w http.ResponseWriter, err error) {
  var errno int
  switch err.(type) {
  case *notFoundError:
    errno = http.StatusNotFound
  default:
    errno = http.StatusInternalServerError
  }

  w.WriteHeader(errno)
  errorTemplate.Execute(w, err)
}

func root(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    if path != "/" {
      errorHandler(w, &notFoundError{})
    } else {
      handler(w, r)
    }
  }
}

func signinRequired(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    token := getCookie(r, "access_token")
    if token == "" {
      http.Redirect(w, r, "/signin", http.StatusFound)
    } else {
      handler(w, r)
    }
  }
}

func partial(handler func(string, http.ResponseWriter, *http.Request), handlerType string) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    handler(handlerType, w, r)
  }
}

func buildAnacondaApiFromRequest(r *http.Request) *anaconda.TwitterApi {
  token := getCookie(r, "access_token")
  secret := getCookie(r, "access_token_secret")
  api := anaconda.NewTwitterApi(token, secret)
  api.ReturnRateLimitError(true)

  return api
}

func buildTimelineRequestValues(r *http.Request) url.Values {
  values := url.Values{}

  maxId, err := strconv.ParseInt(r.FormValue("max_id"), 10, 64)
  if err == nil {
    values.Add("max_id", strconv.FormatInt(maxId - 1, 10))
  }
  values.Add("count", strconv.Itoa(RetrieveCount))

  return values
}

func tweetTextHtml(t *anaconda.Tweet) string {
  html := t.Text

  for _, url := range t.Entities.Urls {
    htmlA := fmt.Sprintf(`<a href="%s">%s</a>`, url.Expanded_url, url.Display_url)
    html = strings.Replace(html, url.Url, htmlA, -1)
  }

  for _, user := range t.Entities.User_mentions {
    htmlA := fmt.Sprintf(`<a href="/user/%s">@%s</a>`, user.Screen_name, user.Screen_name)
    html = strings.Replace(html, "@" + user.Screen_name, htmlA, -1)
  }

  for _, media := range t.Entities.Media {
    htmlA := fmt.Sprintf(`<a href="%s">%s</a>`, media.Expanded_url, media.Display_url)
    html = strings.Replace(html, media.Url, htmlA, -1)
  }

  return html
}

func convertToUserView(u *anaconda.User) *userView {
  user := new(userView)
  user.ProfileImageUrl = u.ProfileImageURL
  user.ScreenName = u.ScreenName
  user.Name = u.Name
  user.Description = u.Description
  user.FriendsCount = u.FriendsCount
  user.FollowersCount = u.FollowersCount
  user.StatusesCount = u.StatusesCount
  user.Location = u.Location
  user.Following = u.Following
  createdAt, _ := time.Parse(time.RubyDate, u.CreatedAt)
  user.CreatedAt = timeToReadableString(createdAt)
  return user
}

func convertToTweetView(t *anaconda.Tweet, screenName string, showOperator bool) *tweetView {
  tweet := new(tweetView)
  tweet.ScreenName = t.User.ScreenName
  tweet.Name = t.User.Name
  tweet.InReplyToStatusId = t.InReplyToStatusID
  tweet.Id = t.Id
  tweet.ProfileImageUrl = t.User.ProfileImageURL
  tweet.Text = template.HTML(tweetTextHtml(t))
  tweet.Source = template.HTML(t.Source)
  if screenName == t.User.ScreenName {
    tweet.ShowRemove = true
  } else {
    tweet.ShowRemove = false
  }
  tweet.ShowOperator = showOperator
  createdAt, _ := t.CreatedAtTime()
  tweet.CreateTime = timeToReadableString(createdAt)
  return tweet
}

func timeToReadableString(t time.Time) string {
  d := time.Since(t)
  if t.Year() != time.Now().Year() {
    return t.Format("2006-01-02")
  } else if d.Hours() > 24 {
    return t.Format("01-02")
  } else if d.Hours() > 1 {
    return fmt.Sprintf("%dh ago", int(d.Hours()))
  } else if d.Minutes() > 1 {
    return fmt.Sprintf("%dm ago", int(d.Minutes()))
  } else {
    return fmt.Sprintf("%ds ago", int(d.Seconds()))
  }
}

func setCookie(w http.ResponseWriter, key string, value string) {
  expires := time.Date(2099, time.November, 10, 23, 0, 0, 0, time.UTC)
  http.SetCookie(w, &http.Cookie{
    Name: key,
    Value: value,
    Expires: expires,
    Path: "/",
    HttpOnly: true })
}

func getCookie(r *http.Request, key string) string {
  cookie, err := r.Cookie(key)
  if err != nil {
    return ""
  } else {
    return cookie.Value
  }
}