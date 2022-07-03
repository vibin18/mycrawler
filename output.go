package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types/events"
	"strings"
	"time"
)

func PrettyPrint(evt *events.Presence, bot *discordgo.Session, users []JIDWithName) {
	var actuallUser string = ""
	var status string = ""
	IndLoc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}
	//	GerLoc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	IndTimeNow := time.Now().In(IndLoc).Format("Mon, 01/02/06, 03:04PM")

	//	GerTimeNow := time.Now().In(GerLoc).Format("2006-01-02 15:04:05")
	for _, user := range users {
		if strings.Contains(fmt.Sprint(evt.From), user.User) {
			actuallUser = fmt.Sprint(user.Name)
		}
		if evt.Unavailable == false {
			status = "online"
		} else {
			status = "offline"
		}

	}
	statusMessage := fmt.Sprintf("User %v is %v at Indian Time %v\n", actuallUser, status, IndTimeNow)

	fmt.Printf(statusMessage)
	_, er := bot.ChannelMessageSend(discordGid, statusMessage)
	if er != nil {
		log.Panicf(er.Error())
		return
	}
}
