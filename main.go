package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	BotToken     = os.Getenv("BOT_TOKEN")
	SearchSecret = os.Getenv("SEARCH_SECRET")
	CX           = os.Getenv("CX")
)

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	s.AddHandler(upHander)
	s.AddHandler(searchHander)

	err := s.Open()
	if err != nil {
		log.Fatalf("error opening the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("gracefully shutdowning")
}

func upHander(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Bot is up!")
}

var prefix = regexp.MustCompile("^(!i|！い)")

func searchHander(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !prefix.Match([]byte(m.Content)) {
		return
	}
	if m.Author.ID == s.State.User.ID {
		return
	}

	searchWord := prefix.ReplaceAllString(m.Content, "")
	searchWord = strings.TrimSpace(searchWord)

	if searchWord == "" {
		log.Printf("search query is empty")
		return
	}

	imageUrl, err := search(searchWord)
	if err != nil {
		log.Printf("error searching image: %v", err)

		message := &discordgo.MessageEmbed{
			Title:       "エラー",
			Description: err.Error(),
			Color:       0xffd700,
		}
		if _, err := s.ChannelMessageSendEmbed(m.ChannelID, message); err != nil {
			log.Printf("error sending error message: %v", err)
		}

		return
	}

	message := &discordgo.MessageEmbed{
		Title: searchWord,
		Image: &discordgo.MessageEmbedImage{
			URL: imageUrl,
		},
		Color: 0x0095d9,
	}
	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, message); err != nil {
		log.Printf("error sending result message: %v", err)
	}
}

func search(query string) (string, error) {
	u, _ := url.Parse("https://www.googleapis.com/customsearch/v1")
	q := url.Values{}
	q.Set("key", SearchSecret)
	q.Set("cx", CX)
	q.Set("num", "1")
	q.Set("searchType", "image")
	q.Set("q", query)
	u.RawQuery = q.Encode()
	println(u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("error calling search api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received not OK response from search API: %s", resp.Status)
	}
	var body SearchResultJson
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Items) == 0 {
		return "", fmt.Errorf("search result is empty")
	}
	return body.Items[0].Link, nil
}

type SearchResultJson struct {
	Items []struct {
		Kind        string `json:"kind"`
		Title       string `json:"title"`
		Htmltitle   string `json:"htmlTitle"`
		Link        string `json:"link"`
		Displaylink string `json:"displayLink"`
		Snippet     string `json:"snippet"`
		Htmlsnippet string `json:"htmlSnippet"`
		Mime        string `json:"mime"`
		Fileformat  string `json:"fileFormat"`
		Image       struct {
			Contextlink     string `json:"contextLink"`
			Height          int    `json:"height"`
			Width           int    `json:"width"`
			Bytesize        int    `json:"byteSize"`
			Thumbnaillink   string `json:"thumbnailLink"`
			Thumbnailheight int    `json:"thumbnailHeight"`
			Thumbnailwidth  int    `json:"thumbnailWidth"`
		} `json:"image"`
	} `json:"items"`
}
