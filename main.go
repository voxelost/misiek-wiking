package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func replaceChar(str string, c byte, idx int) string {
	postfix := ""

	if idx < len(str) {
		postfix = str[idx+1:]
	}

	return fmt.Sprintf("%s%c%s", str[:idx], c, postfix)
}

func VikingifyString(s string) string {
	replaceMap := map[byte]byte{
		'a': 'å',
		'o': 'ø',
		'A': 'Å',
		'O': 'Ø',
	}

	regex := regexp.MustCompile(`(<:.+:\d{18}>)|((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)`)
	urlBoundaries := regex.FindAllStringIndex(s, -1)

	for i := 0; i < len(s); i++ {
		if !func() bool {
			for _, rng := range urlBoundaries {
				if i >= rng[0] && i < rng[1] {
					return false
				}
			}
			return true
		}() {
			continue
		}

		for k, v := range replaceMap {
			if s[i] == k {
				s = replaceChar(s, v, i)
				continue
			}
		}
	}

	return s
}

func AwaitOSInterrupt() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func main() {
	fstream, err := ioutil.ReadFile("config.json")
	CheckErr(err)

	config := make(map[string][]string)
	err = json.Unmarshal(fstream, &config)
	CheckErr(err)

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	CheckErr(err)

	defer dg.Close()

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !func() bool {
			for _, us := range config["known_ids"] {
				if m.Author.ID == us {
					return true
				}
			}
			return false
		}() {
			return
		}

		newMessage := discordgo.MessageSend{
			Content:   VikingifyString(m.Content),
			Reference: m.MessageReference,
		}

		for _, a := range m.Attachments {
			resp, err := http.Get(a.URL)
			if err != nil {
				continue
			}

			newMessage.Files = append(newMessage.Files, &discordgo.File{
				Name:   a.Filename,
				Reader: resp.Body,
			})
		}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &newMessage)
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
