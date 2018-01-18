package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ashwanthkumar/slack-go-webhook"

	"github.com/dghubble/sling"
	upload "github.com/jawher/sling.upload"
	"gopkg.in/yaml.v2"
)

type Config struct {
	HockeyAppId      string `yaml:"hockey_app_id"`
	AppFolder        string `yaml:"app_folder"`
	BuildsPath       string `yaml:"builds_path"`
	SlackChannelName string `yaml:"slack_channel_name"`
}

type UploadResult struct {
	Title        string `json:"title"`
	Verion       string `json:"version"`
	ShortVersion string `json:"shortversion"`
	ConfigUrl    string `json:"config_url"`
	PublicUrl    string `json:"public_url"`
}

func (c *Config) getConf() *Config {

	yamlFile, err := ioutil.ReadFile(".puck.yml")
	if err != nil {
		log.Printf(".puck.yml not found: #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Printf("Invalid .puck.yml: #%v ", err)
	}

	return c
}

func uploadToHockeyApp(c Config) UploadResult {
	buildType := os.Args[1]
	hockeyToken := os.Getenv("HOCKEY_APP_TOKEN")
	uploadUrl := "https://rink.hockeyapp.net/api/2/apps/" + c.HockeyAppId + "/app_versions/upload"
	apkPath := c.AppFolder + "/" + c.BuildsPath + buildType + "/" + c.AppFolder + "-" + buildType + ".apk"

	var uploadResult UploadResult
	_, err := sling.New().Set("X-HockeyAppToken", hockeyToken).Post(uploadUrl).BodyProvider(
		upload.New(
			upload.Part{
				Name:     "ipa",
				FileName: "app.apk",
				Content:  upload.File(apkPath),
			},
			upload.Part{
				Name: "status", Content: upload.String("2"),
			},
		),
	).ReceiveSuccess(&uploadResult)

	if err != nil {
		panic(err)
	}
	return uploadResult
}

func notifyBySlack(c Config, u UploadResult) {
	webhookUrl := os.Getenv("SLACK_WEBHOOK_URL")
	color := "#36a64f"

	author := "Build passed"
	authorUrl := u.PublicUrl

	project := u.Title
	projectUrl := u.ConfigUrl
	fallback := project + ": new build"
	attach := slack.Attachment{
		Fallback:   &fallback,
		Color:      &color,
		AuthorName: &author,
		AuthorLink: &authorUrl,
		Title:      &project,
		TitleLink:  &projectUrl,
	}
	attach.
		AddField(slack.Field{
			Title: "Version name",
			Value: u.ShortVersion,
			Short: true,
		}).
		AddField(slack.Field{
			Title: "Version code",
			Value: u.Verion,
			Short: true,
		})

	payload := slack.Payload{
		Username:    "HockeyApp",
		Channel:     c.SlackChannelName,
		IconEmoji:   ":calling:",
		Attachments: []slack.Attachment{attach},
	}
	err := slack.Send(webhookUrl, "", payload)
	if len(err) > 0 {
		panic(err)
	}
}

func main() {
	var c Config
	c.getConf()

	result := uploadToHockeyApp(c)
	notifyBySlack(c, result)
}
