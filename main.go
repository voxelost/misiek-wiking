package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func VikingifyString(s string) string {
	// ø i zamiast a jest å
	return strings.ReplaceAll(strings.ReplaceAll(s, "a", "å"), "o", "ø")
}

func AwaitOSInterrupt() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func main() {
	fstream, err := ioutil.ReadFile("config.json")
	CheckErr(err)

	config := make(map[string]string)
	err = json.Unmarshal(fstream, &config)
	CheckErr(err)

	dg, err := discordgo.New("Bot " + config["token"])
	CheckErr(err)

	defer dg.Close()

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID != config["misiek_id_1"] && m.Author.ID != config["misiek_id_2"] {
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, VikingifyString(m.Content))
		if err != nil {
			log.Println(err)
			return
		}

		err = dg.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			log.Println(err)
			return
		}
	})

	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	dg.Identify.Presence.Game = discordgo.Activity{
		Type: discordgo.ActivityTypeListening,
		Name: "misiek",
	}

	CheckErr(dg.Open())
	AwaitOSInterrupt()
}