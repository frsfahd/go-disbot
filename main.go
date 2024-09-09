package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dotenv-org/godotenvvault"
)

type quote struct {
	Id     int    `json:"id"`
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

type challenge struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type challenges struct {
	Challenges []challenge `json:"`
}

var listOfChallenges = &challenges{}

func init() {
	if err := godotenvvault.Load(); err != nil {
		slog.Error("failed to load env variables", err)
	}
}

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN") // Replace with your actual bot token

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("Error creating Discord session:", err)
		return
	}

	dg.AddHandler(messageCreatedHandler)

	err = dg.Open()
	if err != nil {
		slog.Error("Error opening connection:", err)
		return
	}
	defer dg.Close()

	slog.Info("Bot is now running. Press Ctrl+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func messageCreatedHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	switch m.Content {
	case "hello":
		greetings(s, m)
		return
	case "!quote":
		generateQuote(s, m)
		return
	case "!challenge":
		generateChallenge(s, m)
		return
	case "!list":
		listAllChallenge(s, m)
	}

	if strings.HasPrefix(m.Content, "!add") {
		addNewChallenge(s, m)
	}
}

func greetings(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Your message handling logic goes here
	username := m.Author.Mention()
	s.ChannelMessageSend(m.ChannelID, "Hello, "+username)

}

func generateQuote(s *discordgo.Session, m *discordgo.MessageCreate) {
	client := http.Client{Timeout: 10 * time.Second}
	uri := "https://dummyjson.com/quotes/random"

	res, err := client.Get(uri)
	if err != nil {
		slog.Error("http.Get() failed with ", err)
	}

	if res.StatusCode != http.StatusOK {
		s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
	}

	dec := json.NewDecoder(res.Body)
	var body quote
	err = dec.Decode(&body)
	if err != nil {
		slog.Error("data.Decode() failed with ", err)
	}

	s.ChannelMessageSend(m.ChannelID, body.Quote)

}

func generateChallenge(s *discordgo.Session, m *discordgo.MessageCreate) {
	f, err := os.Open("challenges.json")
	if err != nil {
		slog.Error("os.Open() failed with ", err)
	}
	defer f.Close()

	var chals challenges
	dec := json.NewDecoder(f)
	err = dec.Decode(&chals)
	if err != nil {
		slog.Error("data.Decode() failed with ", err)
	}

	chal := chals.Challenges[rand.Intn(len(chals.Challenges))]

	s.ChannelMessageSend(m.ChannelID, chal.Name+":"+chal.Url)
}

func addNewChallenge(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := strings.Split(m.Content, " ")
	client := http.Client{Timeout: 10 * time.Second}
	uri := message[1]

	res, err := client.Get(uri)
	if err != nil {
		slog.Error("http.Get() failed with ", err)
	}

	if res.StatusCode != http.StatusOK {
		s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
	}

	htmlByte, _ := io.ReadAll(res.Body)
	// fmt.Println(string(htmlByte))
	title, err := extractTitle(bytes.NewReader(htmlByte))
	if err != nil {
		slog.Error("extractTitle() failed with ", err)
	}

	title = strings.Split(title, "|")[0]

	newChallenge := challenge{Name: title, Url: uri}
	// updatedData, err := json.MarshalIndent(newChallenge, "", " ")
	// if err != nil {
	// 	slog.Error("json.MarshalIndent() failed with ", err)
	// }

	// f, err := os.OpenFile("challenges.json", os.O_APPEND, 0666)
	// if err != nil {
	// 	slog.Error("os.OpenFile() failed with ", err)
	// }

	// _, err = f.Write(updatedData)
	// if err != nil {
	// 	slog.Error("f.Write() failed with ", err)
	// }

	err = updateDataset("challenges.json", newChallenge)
	errCheck("updateDatase()", err)

	// listOfChallenges.Challenges = append(listOfChallenges.Challenges, newChallenge)
	s.ChannelMessageSend(m.ChannelID, "Added: "+title+":"+uri)

}

func listAllChallenge(s *discordgo.Session, m *discordgo.MessageCreate) {
	list := ""
	for _, v := range listOfChallenges.Challenges {
		list += v.Name + ":" + v.Url + "\n"
	}

	s.ChannelMessageSend(m.ChannelID, list)
}
