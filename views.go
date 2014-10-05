package main

import (
  "html/template"
)

type timelineView struct {
  Title string
  ScreenName string
  PreviousPage int
  NextPage int
  SinceId int64
  User *userView
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

var timelineTemplate = template.Must(template.ParseFiles(
    "templates/madoka.tmpl",
    "templates/menu.tmpl",
    "templates/userinfo.tmpl",
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

var signInTemplate = template.Must(template.ParseFiles(
    "templates/signin.tmpl",
    "templates/head.tmpl"))
