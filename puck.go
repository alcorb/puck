package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dghubble/sling"
	upload "github.com/jawher/sling.upload"
	"gopkg.in/yaml.v2"
)

type Config struct {
	HockeyAppId string `yaml:"hockey_app_id"`
	AppFolder   string `yaml:"app_folder"`
	BuildsPath  string `yaml:"builds_path"`
}

type UploadResult struct {
	Verion       string `json:"version"`
	ShortVersion string `json:"shortversion"`
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

func main() {
	var c Config
	c.getConf()

	//slackToken := os.Getenv("SLACK_TOKEN")
	result := uploadToHockeyApp(c)
	fmt.Println(result)
}
