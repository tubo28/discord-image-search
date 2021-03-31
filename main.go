package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	BotToken = flag.String("token", "", "Bot access token")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	prefix := regexp.MustCompile("^(!i|！い)")
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !prefix.Match([]byte(m.Content)) {
			return
		}
		if m.Author.ID == s.State.User.ID {
			return
		}
		word := prefix.ReplaceAllString(m.Content, "")
		s.ChannelMessageSend(m.ChannelID, word)
		f, err := os.Open("test.jpg")
		if err != nil {
			log.Printf("Cannot open image: %v", err)
			return
		}
		defer f.Close()
		if _, err := s.ChannelFileSend(m.ChannelID, "test.jpg", f); err != nil {
			log.Printf("Cannot upload image: %v", err)
		}
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutdowning")
}
