package main

import (
	"flag"
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
	SlackChannelName string `yaml:"slack_channel_name"`
	ApkPath			 string `yaml:"apk_path"`
	BuildType      	 string `yaml:"build_type"`
	DescriptionPath  string `yaml:"description_path"`
}

type UploadResult struct {
	Title        string `json:"title"`
	Verion       string `json:"version"`
	ShortVersion string `json:"shortversion"`
	ConfigUrl    string `json:"config_url"`
	PublicUrl    string `json:"public_url"`
}

func (c *Config) getConf(configPath *string) *Config {

	yamlFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Printf(*configPath+" not found: #%v ", err)
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Printf("Invalid "+(*configPath)+": #%v ", err)
		panic(err)
	}

	return c
}
func uploadToHockeyApp(c Config) UploadResult {
	hockeyToken := os.Getenv("HOCKEY_APP_TOKEN")
	uploadUrl := "https://rink.hockeyapp.net/api/2/apps/" + c.HockeyAppId + "/app_versions/upload"
	apkPath := c.ApkPath
	descriptionPath := c.DescriptionPath

	var descriptionPart upload.Part
	if descriptionPath != "" {
		descriptionPart := upload.Part{
			Name:		"notes",
			FileName:	"description.txt",
			Content: upload.File(descriptionPath),
		}
	}					

	var uploadResult UploadResult
	_, err := sling.New().
					Set("X-HockeyAppToken", hockeyToken).
					Post(uploadUrl).
					BodyProvider(
						upload.New(
							upload.Part{
								Name:     "ipa",
								FileName: "app.apk",
								Content:  upload.File(apkPath),
							},
						upload.Part{
							Name: "status", Content: upload.String("2"),
						},
						descriptionPart,
						upload.Part{
							Name:		"notes_type",
							Content:	upload.String("1"),
						},
						),
					).
					ReceiveSuccess(&uploadResult)

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

	project := u.Title + " [" + c.BuildType + "]"
	projectUrl := u.ConfigUrl
	fallback := project + ": new build"
	//hockeyInstall := "Install HockeyApp: https://rink.hockeyapp.net/apps/0873e2b98ad046a92c170a243a8515f6"

	attach := slack.Attachment{
		//PreText:    &hockeyInstall,
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
	configPath := flag.String("config", ".puck.yml", "path to config file")
	flag.Parse()

	var c Config
	c.getConf(configPath)

	result := uploadToHockeyApp(c)
	notifyBySlack(c, result)
}
