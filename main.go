package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

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

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "i",
			Description: "画像検索",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search-word",
					Description: "この単語で検索します",
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"i": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			f, err := os.Open("test.jpg")
			defer f.Close()
			if err != nil {
				log.Printf("Cannot upload image: %v", err)
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				// Ignore type for now, we'll discuss them in "responses" part
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: "hello",
				},
			})
			s.ChannelFileSend(i.ChannelID, "test.jpg", f)
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	guilds := s.State.Guilds
	var cmds []*discordgo.ApplicationCommand

	for _, g := range guilds {
		for _, v := range commands {
			log.Printf("Create command '%s' guild %s", v.Name, g.ID)
			cmd, err := s.ApplicationCommandCreate(s.State.User.ID, g.ID, v)
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			}
			cmds = append(cmds, cmd)
		}
	}
	defer func() {
		for _, g := range guilds {
			for _, v := range cmds {
				log.Printf("Delete command '%v' guild %v", v.Name, g.ID)
				if err := s.ApplicationCommandDelete(s.State.User.ID, g.ID, v.ID); err != nil {
					log.Printf("Cannot delete '%v' guild %v command: %v", v.Name, g.ID, err)
				}
			}
		}
	}()

	defer s.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutdowning")
}
