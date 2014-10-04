package twidoka

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