package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	Challenges []challenge `json:challenges"`
}

var listOfChallenges *challenges

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

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		var messageOutput string
		var err error

		switch m.Content {
		case "hello":
			messageOutput, err = greetings(s, m)
		case "!quote":
			messageOutput, err = generateQuote(s, m)
		case "!challenge":
			messageOutput, err = generateChallenge(s, m)
		case "!list":
			messageOutput, err = listAllChallenge(s, m)
		}

		if strings.HasPrefix(m.Content, "!add") {
			messageOutput, err = addNewChallenge(s, m)
		}

		if err != nil {
			slog.Error("messageHandler failed with, ", err)
		}
		s.ChannelMessageSend(m.ChannelID, messageOutput)
	})

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
