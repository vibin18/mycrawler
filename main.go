package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	flags "github.com/jessevdk/go-flags"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type opts struct {
	DiscordToken     string `           long:"token"      env:"DISCORD_TOKEN"  description:"Discord Bot token" required:"true"`
	DiscordChannelId string `           long:"id"      env:"DISCORD_CHANNEL"  description:"Discord Channel ID" required:"true"`
	UserFile         string `           long:"file"      env:"USER_FILE"  description:"User file" required:"true"`
}

type JIDWithName struct {
	types.JID
	Name string
}

func user2JIDWithName(user fileUser) *JIDWithName {
	var jidObj JIDWithName
	jidObj.JID.User = fmt.Sprint(user.Number)
	jidObj.JID.Server = "s.whatsapp.net"
	jidObj.Name = user.Name
	return &jidObj
}

var (
	argparser    *flags.Parser
	arg          opts
	discordToken string
	discordGid   string
	userFile     string
	bot          *discordgo.Session
	whatUsers    []JIDWithName
)

func initArgparser() {
	argparser = flags.NewParser(&arg, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

func eventHandler(evt interface{}) {

	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())

	case *events.Presence:

		myPrint()
		PrettyPrint(v, bot, whatUsers)
		myPrint()
	}
}

func myPrint() {
	fmt.Println("---------")
}

func main() {
	initArgparser()
	discordToken = arg.DiscordToken
	discordGid = arg.DiscordChannelId
	userFile = arg.UserFile

	var err error
	bot, err = discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Panicf("Error: %v", err)

		f := fileData{"user.yaml"}
		usersList := f.GetUsers()
		whatUsers := []JIDWithName{}
		for _, items := range usersList.Users {
			u1 := user2JIDWithName(items)
			whatUsers = append(whatUsers, *u1)
		}

		dbLog := waLog.Stdout("Database", "DEBUG", true)
		// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
		container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
		if err != nil {
			log.Panicf("Error: %v", err)

		}
		deviceStore, err := container.GetFirstDevice()
		if err != nil {
			log.Panicf("Error: %v", err)
		}
		clientLog := waLog.Stdout("Client", "DEBUG", true)
		client := whatsmeow.NewClient(deviceStore, clientLog)
		client.AddEventHandler(eventHandler)
		container.GetAllDevices()

		if client.Store.ID == nil {
			// No ID stored, new login
			qrChan, _ := client.GetQRChannel(context.Background())
			err = client.Connect()
			if err != nil {
				panic(err)
			}
			for evt := range qrChan {
				if evt.Event == "code" {
					// Render the QR code here
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
					// Use --> https://www.webtoolkitonline.com/qrcode-generator.html
					fmt.Println("QR code:", evt.Code)
				} else {
					fmt.Println("Login event:", evt.Event)
				}
			}
		} else {
			// Already logged in, just connect
			err = client.Connect()
			if err != nil {
				log.Panicf("Error: %v", err)
			}
		}

		const (
			PresenceAvailable types.Presence = "available"
		)

		wg := sync.WaitGroup{}
		s := 0
		for s == 0 {
			fmt.Printf("%v th status checking", s)
			for _, user := range whatUsers {
				wg.Add(1)
				go func() {
					err := client.SendPresence(PresenceAvailable)
					if err != nil {
						log.Panicf("Error: %v", err)
					}
					err = client.SubscribePresence(user.JID)
					if err != nil {
						log.Panicf("Error: %v", err)
					}
					wg.Done()
				}()
				wg.Wait()
			}
			time.Sleep(3 * time.Second)
		}

		// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		client.Disconnect()
	}
}
