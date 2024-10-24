package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Discord struct {
	webhookUrl string
}

func (d *Discord) sendMessage(message string) error {
	// Encode the message to JSON
	type DiscordReq struct {
		Username string `json:"username"`
		Content  string `json:"content"`
	}
	json, err := json.Marshal(DiscordReq{
		Username: "Factrio Server Watcher", Content: message,
	})
	if err != nil {
		return err
	}

	// Post the JSON
	_, err = http.Post(d.webhookUrl, "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}

	return nil
}

func doReadCheckNotifyLoop(d *Discord, src io.Reader) error {
	// Regex for check
	regex := regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} \[(JOIN|LEAVE)\] (.+) (?:joined the game|left the game)$`)

	// Read each line and check&notify
	reader := bufio.NewReader(src)
	for {
		// Read
		line, err := reader.ReadString('\n')
		if line == "" || err == io.EOF {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return err
		}
		fmt.Print(line)
		line = strings.TrimSpace(line)

		// Check
		if !regex.MatchString(line) {
			continue
		}

		// Notify
		err = d.sendMessage(line)
		if err != nil {
			return err
		}
	}
}

func main() {
	webhookUrl, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
	if !ok {
		log.Fatal("Set environment variable DISCORD_WEBHOOK_URL")
	}

	discord := &Discord{webhookUrl}
	err := doReadCheckNotifyLoop(discord, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
}
