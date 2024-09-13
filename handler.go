package main

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func greetings(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	// Your message handling logic goes here
	username := m.Author.Mention()
	if username == "" {
		return "", errors.New("empty usernmae")
	}

	return "Hello, " + username, nil

}

func generateQuote(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	uri := "https://dummyjson.com/quotes/random"

	res, err := client.Get(uri)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		// s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		return "", err
	}

	// dec := json.NewDecoder(res.Body)
	var body quote

	err = decodeFromReader(res.Body, &body)
	if err != nil {
		return "", err
	}

	return body.Quote, nil

}

func generateChallenge(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	f, err := os.ReadFile("challenges.json")
	if err != nil {
		return "", err
	}

	err = decodeFromReader(bytes.NewReader(f), &listOfChallenges)
	if err != nil {
		return "", err
	}

	chal := listOfChallenges.Challenges[rand.Intn(len(listOfChallenges.Challenges))]

	return chal.Name + ":" + chal.Url, nil

}

func addNewChallenge(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	slog.Info("addNewChallenge()")

	message := strings.Split(m.Content, " ")
	client := http.Client{Timeout: 10 * time.Second}
	uri := message[1]

	res, err := client.Get(uri)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	htmlByte, _ := io.ReadAll(res.Body)
	// fmt.Println(string(htmlByte))
	title, err := extractTitle(bytes.NewReader(htmlByte))
	if err != nil {
		return "", err
	}

	title = strings.Split(title, "|")[0]

	newChallenge := challenge{Name: title, Url: uri}

	err = updateDataset("challenges.json", newChallenge)
	if err != nil {
		return "", err
	}

	// listOfChallenges.Challenges = append(listOfChallenges.Challenges, newChallenge)
	return "Added: " + title + ":" + uri, nil

}

func listAllChallenge(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	slog.Info("listAllChallenge()")
	var buf bytes.Buffer
	defer buf.Reset()

	f, err := os.ReadFile("challenges.json")
	if err != nil {
		return "", err
	}

	err = decodeFromReader(bytes.NewReader(f), &listOfChallenges)
	if err != nil {
		return "", err
	}

	for _, v := range listOfChallenges.Challenges {
		buf.WriteString(v.Name + ":" + v.Url + "\n")
	}

	return buf.String(), nil
}
