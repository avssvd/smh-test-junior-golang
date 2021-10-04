package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getUserKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Check IP"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Get list of checked IPs"),
			tgbotapi.NewKeyboardButton("Get list of checked IPs results"),
		),
	)
}

func getAdminKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Send broadcast message"),
			tgbotapi.NewKeyboardButton("Get list of user's checked IPs"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Add new admin"),
			tgbotapi.NewKeyboardButton("Remove admin"),
		),
	)
}

func tgBot(env *Env) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Error(err)
		return
	}

	sendSafe := func(c tgbotapi.Chattable) {
		for _, err := bot.Send(c); err != nil; {
			log.Error(err)
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Error(err)
		return
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

UpdateLoop:
	for update := range updates {
		// Check is user in DB
		user, err := env.users.Get(update.Message.From.ID)
		switch {
		case errors.Is(err, ErrUserNotFound):
			// If not exist -> create
			user = getNewUser(update.Message.From)
			if err := env.users.Insert(user); err != nil {
				log.Error(err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong")
				sendSafe(msg)
				continue UpdateLoop
			}

		case err != nil:
			log.Error(err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong")
			sendSafe(msg)
			continue UpdateLoop

		default:
			// If exist -> update
			if err := env.users.UpdateInfo(user, getNewUser(update.Message.From)); err != nil {
				log.Error(err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong")
				sendSafe(msg)
				continue UpdateLoop
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch user.IsAdmin {
		case true:
			tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = getAdminKeyboard()
			switch {
			case update.Message != nil && update.Message.IsCommand():
				switch update.Message.Command() {

				case "start":
					msg.Text = "Hi. Use the keyboard for actions."

				default:
					msg.Text = "I don't know that command"
				}
			case update.Message != nil && update.Message.ReplyToMessage == nil:
				switch update.Message.Text {
				case "Send broadcast message":
					msg.ParseMode = "html"
					msg.Text = "Send broadcast message\n" +
						"Reply to this message with broadcast message text"

				case "Get list of user's checked IPs":
					msg.ParseMode = "html"
					msg.Text = "Get list of user's checked IPs\n" +
						"Reply to this message with user Telegram ID"

				case "Add new admin":
					msg.ParseMode = "html"
					msg.Text = "Add new admin\n" +
						"Reply to this message with new admin Telegram ID\n" +
						"NB: new admin should have a dialogue with me!"

				case "Remove admin":
					msg.ParseMode = "html"
					msg.Text = "Remove admin\n" +
						"Reply to this message with deprecated admin Telegram ID"

				}
			case update.Message != nil && update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == bot.Self.ID:
				switch strings.FieldsFunc(update.Message.ReplyToMessage.Text, func(r rune) bool { return r == '\n' })[0] {
				case "Send broadcast message":
					recipients, err := env.users.List()
					if err != nil {
						log.Error(err)
					}
					for _, recipient := range recipients {
						broadcastMsg := tgbotapi.NewMessage(int64(recipient.TgID), "")
						broadcastMsg.ParseMode = "html"
						broadcastMsg.Text = update.Message.Text
						sendSafe(broadcastMsg)
					}

				case "Get list of user's checked IPs":
					msg.ParseMode = "html"
					userTgID, err := strconv.Atoi(update.Message.Text)
					if err != nil {
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("%v is invalid Telegram ID value\nShould be unsigned integer", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					}
					msg.Text = "Checked IPs:"
					ipChecks, err := env.ipChecks.ListByTgID(userTgID, true)
					switch {
					case errors.Is(err, ErrUserNotFound):
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("User with Telegram ID %v not found", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					case err != nil:
						log.Error(err)
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong\nTry again later")
						sendSafe(errMsg)
						continue UpdateLoop
					}
					for _, ipCheck := range ipChecks {
						msg.Text += "\n" + ipCheck.IP
					}

				case "Add new admin":
					msg.ParseMode = "html"
					userTgID, err := strconv.Atoi(update.Message.Text)
					if err != nil {
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("%v is invalid Telegram ID value\nShould be unsigned integer", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					}

					err = env.users.SetAdminStatus(userTgID, true)
					switch {
					case errors.Is(err, ErrUserNotFound):
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("User with Telegram ID %v not found", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					case err != nil:
						log.Error(err)
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong\nTry again later")
						sendSafe(errMsg)
						continue UpdateLoop
					}
					msg.Text = "Success"

				case "Remove admin":
					msg.ParseMode = "html"
					userTgID, err := strconv.Atoi(update.Message.Text)
					if err != nil {
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("%v is invalid Telegram ID value\nShould be unsigned integer", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					}

					err = env.users.SetAdminStatus(userTgID, false)
					switch {
					case errors.Is(err, ErrUserNotFound):
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("User with Telegram ID %v not found", update.Message.Text))
						sendSafe(errMsg)
						continue UpdateLoop
					case err != nil:
						log.Error(err)
						errMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something goes wrong\nTry again later")
						sendSafe(errMsg)
						continue UpdateLoop
					}
					msg.Text = "Success"
				}
			}

		case false:
			// Update keyboard
			tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = getUserKeyboard()
			switch {
			case update.Message != nil && update.Message.IsCommand():
				switch update.Message.Command() {

				case "start":
					msg.Text = "Hi. Use the keyboard for actions."

				default:
					msg.Text = "I don't know that command"
				}
			case update.Message != nil && update.Message.ReplyToMessage == nil:
				switch update.Message.Text {
				case "Check IP":
					msg.ParseMode = "html"
					msg.Text = "Check IP\n" +
						"Reply to this message with IP address what you want to check\n" +
						"Examples: <pre>8.8.8.8</pre>"

				case "Get list of checked IPs":
					msg.ParseMode = "html"
					msg.Text = "Checked IPs:"
					ipChecks, err := env.ipChecks.ListByTgID(user.TgID, true)
					if err != nil {
						log.Error(err)
					}
					for _, ipCheck := range ipChecks {
						msg.Text += "\n" + ipCheck.IP
					}

				case "Get list of checked IPs results":
					ipChecks, err := env.ipChecks.ListByTgID(user.TgID, true)
					if err != nil {
						log.Error(err)
					}
					for _, ipCheck := range ipChecks {
						ipInfo := IPInfo{}
						byteIPInfo, err := ipCheck.IPInfo.MarshalJSON()
						if err != nil {
							log.Error(err)
							continue
						}
						err = json.Unmarshal(byteIPInfo, &ipInfo)
						if err != nil {
							log.Error(err)
							continue
						}
						ipCheckMsg := tgbotapi.NewMessage(update.Message.Chat.ID, ipInfo.MessageString())
						ipCheckMsg.ParseMode = "html"
						_, err = bot.Send(ipCheckMsg)
						if err != nil {
							log.Error(err)
						}
					}
				}
			case update.Message != nil && update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == bot.Self.ID:
				switch strings.FieldsFunc(update.Message.ReplyToMessage.Text, func(r rune) bool { return r == '\n' })[0] {
				case "Check IP":
					msg.ParseMode = "html"
					ipAddr := net.ParseIP(update.Message.Text)
					if ipAddr.String() != "<nil>" {
						ipInfo, err := getIPInfo(ipAddr)
						if err != nil {
							log.Error(err)
						} else {
							err = env.ipChecks.Insert(&IPCheck{IP: ipAddr.String(), IPInfo: ipInfo.JSONBytes(), UserTgID: user.TgID})
							if err != nil {
								log.Error(err)
							}
							msg.Text = ipInfo.MessageString()
						}
					} else {
						msg.Text = "Check IP" +
							fmt.Sprintf("\n\n<code>%v</code> is not a valid textual representation of an IP address!\n", update.Message.Text) +
							"Try again"
					}
				}
			}
		}

		if msg.Text != "" {
			msg.ReplyToMessageID = update.Message.MessageID
			sendSafe(msg)
		}
	}
}
