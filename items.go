package main

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// добавляем предмет
func (player *Player) EquipItem(item Item) {
	// Проверяем, есть ли предмет в инвентаре
	for i, invItem := range player.Inventory {
		if invItem.ID == item.ID {
			// Удаляем предмет из инвентаря
			player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
			// Добавляем предмет в экипированные
			player.EquippedItems = append(player.EquippedItems, item)
			// Применяем статы предмета к игроку
			player.ApplyItemStats(item)
			break
		}
	}
}

func (player *Player) ApplyItemStats(item Item) {
	switch item.Stat {
	case "HP":
		player.MaxHP += item.Value
		player.HP += item.Value
	case "Attack":
		player.Attack += item.Value
	case "Defense":
		player.Defense += item.Value
	}
}

func showInventory(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	player := a.getPlayer(message.From.ID)
	if player == nil {
		return
	}

	var inventoryText strings.Builder
	inventoryText.WriteString("Ваш инвентарь:\n")
	itemsp := a.PlayerItems(message.From.ID)
	for _, i := range itemsp {
		inventoryText.WriteString(fmt.Sprintf("- %s (%s: +%d)\n", i.Name, i.Stat, i.Value))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, inventoryText.String())
	bot.Send(msg)
}

func handleEquipItem(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	player := a.getPlayer(message.From.ID)
	if player == nil {
		return
	}

	itemName := strings.TrimSpace(strings.TrimPrefix(message.Text, "Надеть"))
	for _, item := range player.Inventory {
		if item.Name == itemName {
			player.EquipItem(item)
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Вы надели предмет: %s!", item.Name))
			bot.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Предмет не найден в инвентаре.")
	bot.Send(msg)
}
