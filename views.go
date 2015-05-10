package main

import (
  "html/template"
)

type timelineView struct {
  Title string
  ScreenName string
  SinceId int64
  User *userView
  Search string
  Referer string
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
  ImageUrl string
  Source template.HTML
  ShowFull bool
  Mentioned []string
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
  Referer string
}

type userView struct {
  ProfileImageUrl string
  ScreenName string
  Name string
  Description string
  FriendsCount int
  FollowersCount int
  StatusesCount int64
  Location string
  Following bool
  CreatedAt string
}

type removeView struct {
  Tweet *tweetView
  Referer string
}

var timelineTemplate = template.Must(template.ParseFiles(
    "templates/madoka.tmpl",
    "templates/menu.tmpl",
    "templates/userinfo.tmpl",
    "templates/tweet_list.tmpl",
    "templates/searchbox.tmpl",
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
    "templates/menu.tmpl",
    "templates/head.tmpl"))

var signInTemplate = template.Must(template.ParseFiles(
    "templates/signin.tmpl",
    "templates/head.tmpl"))

var removeTemplate = template.Must(template.ParseFiles(
    "templates/remove.tmpl",
    "templates/menu.tmpl",
    "templates/tweet.tmpl",
    "templates/head.tmpl"))