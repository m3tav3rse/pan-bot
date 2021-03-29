package main

import (
	"flag"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const helpMessage = `
!help            Na pewno nie wyświetli tej listy komend
	!sjp    <wyraz>  Znaczenie wyrazu z sjp.pl
	!echo   <tekst>  Wypisuje tekst
`

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

func main() {

	if token == "" {
		log.Println("No token provided. Please run:", os.Args[0], "-t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error creating Discord session: ", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	dg.AddHandler(messageReactionAdd)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildEmojis | discordgo.IntentsGuildMessageReactions

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalln("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Pan Bot started.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch {
	case strings.HasPrefix(m.Content, "!echo"):
		s.ChannelMessageSend(m.ChannelID, strings.TrimSpace(m.Content[5:]))
	case strings.HasPrefix(m.Content, "!help"):
		s.ChannelMessageSend(m.ChannelID, helpMessage)
	case strings.HasPrefix(m.Content, "!sjp"):
		message := sjpQuery(strings.TrimSpace(m.Content[4:]))
		s.ChannelMessageSend(m.ChannelID, message)
	case strings.ToLower(m.Content) == "kek",
		strings.ToLower(m.Content) == "lol",
		strings.Contains(strings.ToLower(m.Content), "kekw"):
		e := getEmoji(s, "KEKW", m.GuildID)
		if e != "" {
			s.MessageReactionAdd(m.ChannelID, m.ID, e)
		}
	case strings.ToLower(m.Content) == "idź sobie":
		s.ChannelMessageSendReply(m.ChannelID, "Nie", m.Reference())
	}
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}

	switch {
	case m.Emoji.Name == "KEKW":
		s.MessageReactionAdd(m.ChannelID, m.MessageID, m.Emoji.APIName())
	default:
		s.MessageReactionAdd(m.ChannelID, m.MessageID, m.Emoji.APIName())
	}
}

func getEmoji(s *discordgo.Session, emoji string, guildID string) string {
	list, err := s.GuildEmojis(guildID)
	if err != nil {
		log.Println(err)
	}

	for _, e := range list {
		if e.Name == "KEKW" {
			return e.APIName()
		}
	}
	return ""
}

func sjpQuery(word string) string {
	resp, err := http.Get("https://sjp.pl/" + word)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	parAll, err := regexp.Compile("<p style=\"margin: .5em 0; font: medium/1.4 sans-serif; max-width: 32em; \">.*</p>")
	if err != nil {
		log.Println(err)
	}

	parStart, err := regexp.Compile("<p style=\"margin: .5em 0; font: medium/1.4 sans-serif; max-width: 32em; \">")
	if err != nil {
		log.Println(err)
	}

	parEnd, err := regexp.Compile("</p>")
	if err != nil {
		log.Println(err)
	}

	linebreak, err := regexp.Compile("<br />")
	if err != nil {
		log.Println(err)
	}

	text := parAll.FindAllString(string(body), -1)

	var result string

	for _, x := range text {
		s := html.UnescapeString(x)
		s = linebreak.ReplaceAllString(s, "\n")
		s = parStart.ReplaceAllString(s, "")
		s = parEnd.ReplaceAllString(s, "")
		result += s + "\n"
	}

	result = strings.Trim(result, " \t\n")

	if result == "" {
		return "Nie występuje w słowniku"
	} else {
		return result
	}
}
