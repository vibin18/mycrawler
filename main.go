package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	flags "github.com/jessevdk/go-flags"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type opts struct {
	DiscordToken     string `           long:"token"      env:"DISCORD_TOKEN"  description:"Discord Bot token" required:"true"`
	DiscordChannelId string `           long:"id"      env:"DISCORD_CHANNEL"  description:"Discord Channel ID" required:"true"`
	UserFile         string `           long:"file"      env:"USER_FILE"  description:"user file" required:"true"`
}

var (
	argparser    *flags.Parser
	arg          opts
	discordToken string
	discordGid   string
	userFile     string
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
	bot, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		panic(err)
	}
	f := fileData{"user.yaml"}
	usersList := f.GetUsers()
	Users := []JIDWithName{}
	for _, items := range usersList.Users {
		u1 := user2JIDWithName(items)
		Users = append(Users, *u1)
	}
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())

	case *events.Presence:

		myPrint()
		PrettyPrint(v, bot, Users)
		myPrint()
	}
}

type JIDWithName struct {
	types.JID
	Name string
}

func user2JIDWithName(user fileUser) *JIDWithName {
	var JIDObj JIDWithName
	JIDObj.JID.User = fmt.Sprint(user.Number)
	JIDObj.JID.Server = "s.whatsapp.net"
	JIDObj.Name = user.Name
	return &JIDObj
}

//var users = []JIDWithName{userVibin, userJoseKuttan, userAmma, userJozemon, userNikhil, userHenna, userJoeVakkan, userJoeSeby}

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
func myPrint() {
	fmt.Println("---------")
}

func main() {
	initArgparser()
	discordToken = arg.DiscordToken
	discordGid = arg.DiscordChannelId
	userFile = arg.UserFile

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)

	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
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
			panic(err)
		}
	}

	const (
		PresenceAvailable types.Presence = "available"
	)
	f := fileData{userFile}
	usersList := f.GetUsers()
	Users := []JIDWithName{}
	for _, items := range usersList.Users {
		u1 := user2JIDWithName(items)
		Users = append(Users, *u1)
	}

	wg := sync.WaitGroup{}
	for s := range [100000]int{} {
		fmt.Printf("%v th status checking", s)
		for _, user := range Users {
			wg.Add(1)
			go func() {
				client.SendPresence(PresenceAvailable)
				err := client.SubscribePresence(user.JID)
				if err != nil {
					panic(err)
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
