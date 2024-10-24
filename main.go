package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
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

func doReadCheckNotifyLoop(d *Discord, src io.Reader) {
	// Regex for check
	regex := regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} \[(JOIN|LEAVE)\] (.+) (?:joined the game|left the game)$`)

	// Read each line and check&notify
	reader := bufio.NewReader(src)
	for {
		// Read
		line, err := reader.ReadString('\n')
		if err != nil {
			return
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
			continue
		}
	}
}

func executeFactorio(ctx context.Context, args []string) (*exec.Cmd, io.Reader, error) {
	cmd := exec.CommandContext(ctx, args[1], args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 10 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	cmd.Start()

	return cmd, stdout, nil
}

func doMain() error {
	webhookUrl, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
	if !ok {
		return errors.New("Set environment variable DISCORD_WEBHOOK_URL")
	}
	discord := &Discord{webhookUrl}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd, stdout, err := executeFactorio(ctx, os.Args)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		doReadCheckNotifyLoop(discord, stdout)
	}()

	return cmd.Wait()
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("needs arguments")
	}

	if err := doMain(); err != nil {
		log.Fatal(err)
	}
}
