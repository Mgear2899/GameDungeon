package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func showInventory(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	player := a.getPlayer(message.Chat.ID)
	if player == nil {
		return
	}

	var buttons [][]tgbotapi.InlineKeyboardButton

	itemsp := a.PlayerItems(message.Chat.ID)
	_, itemID := a.PlayerItemCount(int64(player.ID))
	for _, i := range itemsp {
		var btnText string

		isEquipped := false
		for _, id := range itemID {
			if id.itemID == i.ID {
				isEquipped = true
				break
			}
		}

		if isEquipped {
			btnText = fmt.Sprintf("✅%s (+%s %d)", i.Name, i.Stat, i.Value)
		} else {
			btnText = fmt.Sprintf("%s (+%s %d)", i.Name, i.Stat, i.Value)
		}

		callbackData := fmt.Sprintf("equip:%d", i.ID)
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData),
			))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери предмет для экипировки:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	bot.Send(msg)
}

// добавить или снять предмет
func handleEquip(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery, playerID int64) {
	var (
		itemID int
		text   string
		pm     string
		slot   int
	)
	// action_104
	// re := regexp.MustCompile("")
	res := strings.HasPrefix(cq.Data, "action_")
	if res {
		fmt.Println("True", cq.Data)
	} else {
		fmt.Println("False", cq.Data)
	}
	_, err := fmt.Sscanf(cq.Data, "equip:%d", &itemID)
	if err != nil {
		log.Println("Ошибка парсинга:", err)
		return
	}

	count, itemArr := a.PlayerItemCount(playerID)

	isEquipped := false
	for _, id := range itemArr {
		if id.itemID == itemID {
			isEquipped = true
			break
		}
	}

	if isEquipped {
		text = "Вы сняли %s (+%s %d)"
		pm = "-"
		a.DeleteItem(itemID, int(playerID))
	} else {
		// проверим сколько надето
		if count >= 2 {
			msg := tgbotapi.NewMessage(cq.Message.Chat.ID, "Уже надето 2 предмета!")
			bot.Send(msg)
			return
		}

		if len(itemArr) == 0 { // записей не нашли
			slot = 0
		} else if itemArr[0].slot == 0 {
			slot = 1
		} else {
			slot = 0
		}

		a.PlayerItemEquip(int(playerID), itemID, slot)
		text = "Вы надели %s (+%s %d)"
		pm = "+"
	}

	var name, stat string
	var value int
	items := a.PlayerItems(playerID)
	for _, i := range items {
		if i.ID == itemID {
			name = i.Name
			stat = i.Stat
			value = i.Value
			break
		}

	}

	a.RecalcStats(playerID, stat, value, pm)

	showInventory(bot, cq.Message)

	msg := tgbotapi.NewMessage(cq.Message.Chat.ID,
		fmt.Sprintf(text, name, stat, value))
	bot.Send(msg)
}

// newText := ""
// 			if update.CallbackQuery.Data == "opt1" {
// 				newText = "Вы выбрали опцию 1 ✅"
// 			} else {
// 				newText = "Вы выбрали опцию 2 ✅"
// 			}

// 			edit := tgbotapi.NewEditMessageText(
// 				update.CallbackQuery.Message.Chat.ID,
// 				update.CallbackQuery.Message.MessageID,
// 				newText,
// 			)
