package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var webhookURL = os.Getenv("URL")

func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// Discord webhook payload structures
type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type embedFooter struct {
	Text string `json:"text"`
}

type embed struct {
	Title     string       `json:"title"`
	Color     int          `json:"color"`
	Fields    []embedField `json:"fields"`
	Footer    embedFooter  `json:"footer"`
	Timestamp string       `json:"timestamp"`
}

type webhookPayload struct {
	Embeds []embed `json:"embeds"`
}

func sendDiscordEmbed(name, rollNumber string, timestamp time.Time) error {
	// #083173 in decimal
	const embedColor = 0x083173

	payload := webhookPayload{
		Embeds: []embed{
			{
				Title: "✅ Challenge Passed",
				Color: embedColor,
				Fields: []embedField{
					{
						Name:   "👤 Name",
						Value:  name,
						Inline: true,
					},
					{
						Name:   "🎓 Roll Number",
						Value:  rollNumber,
						Inline: true,
					},
				},
				Footer: embedFooter{
					Text: "Docker Workshop Challenge",
				},
				// Discord expects ISO 8601 UTC timestamp
				Timestamp: timestamp.UTC().Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned non-OK status: %s", resp.Status)
	}

	return nil
}

func main() {
	if !isRunningInDocker() {
		fmt.Println("You are not running inside a Docker container.")
		fmt.Println("This application is meant to be run inside a Docker container. Please build the image and run it as a container!")
		return
	}

	if webhookURL == "" {
		fmt.Println("Error: URL environment variable is not set.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter your roll number: ")
	rollNumber, _ := reader.ReadString('\n')
	rollNumber = strings.TrimSpace(rollNumber)

	fmt.Printf("\nHello, %s! Your roll number is %s.\n", name, rollNumber)
	fmt.Println("Welcome to the Docker Workshop!")

	fmt.Println("\nSending your submission to Discord...")
	if err := sendDiscordEmbed(name, rollNumber, time.Now()); err != nil {
		fmt.Printf("Could not send Discord notification :(, error: %v\n", err)
	} else {
		fmt.Println("Submission sent successfully! :)")
	}
}
