package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "modernc.org/sqlite"
)

type Player struct {
	ID            int
	Name          string
	Class         string
	HP            int
	Attack        int
	Defense       int
	Agility       int
	Gold          int
	Stage         int
	XP            int
	Level         int
	Count         int
	MaxHP         int
	Inventory     []Item
	EquippedItems []Item
}

type Item struct {
	ID    int
	Name  string
	Stat  string
	Value int
	Price int
}

type Class struct {
	Name    string
	BaseHP  int
	BaseAtk int
	BaseDef int
	Agility int
	Button  Skils
}

type Skils struct {
	Name      string
	NameSkill string
	Help      string
	BaseHP    int
	BaseAtk   int
	BaseDef   int
	Agility   int
}

type Monster struct {
	Name     string
	HP       int
	AtkPower int
	Def      int
	XP       int
	Gold     int
	IsBoss   bool
}

const XPPerLevel = 180

var counts = make(map[int64]int)
var countPotion = make(map[int64]int)
var MonsterHP = make(map[int64]*stateFight)
var ItemMap = make(map[int64]Item)
var ArrayItemShell = make(map[int64]int)

var a App

func telebot(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	bot.Debug = false
	log.Printf("Авторизовался аккаунт %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	return updates
}

func main() {
	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		log.Panic("ошибка подключения: ", err)
	}

	updates := telebot(bot)

	CraeteDataBase()

	// Инициализация БД
	a.ConnDB()
	defer a.DB.Close()

	// Создание таблиц
	a.createTablesAndDB()
	a.DeleteTime()

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		}

		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	// Извлекаем данные из callback-запроса
	data := callbackQuery.Data
	bot.Request(tgbotapi.NewCallback(callbackQuery.ID, ""))

	// Проверяем, что это за действие
	if strings.HasPrefix(data, "action_") {
		// Обработка действия с предметом
		itemIDStr := strings.TrimPrefix(data, "action_")
		itemID, err := strconv.Atoi(itemIDStr)
		if err != nil {
			log.Printf("Ошибка при преобразовании ID предмета: %v", err)
			return
		}

		// выбираем предмет
		selectItem(bot, callbackQuery, itemID)
	} else if data == "item_sell" {
		// продажа предмета
		sellItems(bot, callbackQuery)
	}

	if len(data) > 0 {
		chatID := callbackQuery.Message.Chat.ID
		handleEquip(bot, callbackQuery, chatID)

	}

	// Отправляем ответ на callback-запрос (чтобы убрать "часики" у кнопки)
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Ошибка при отправке ответа на callback: %v", err)
	}
}

func selectItem(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, itemID int) {
	chatID := callbackQuery.Message.Chat.ID
	ArrayItemShell[chatID] = itemID
	// Обновляем сообщение с кнопками (если нужно)
	editMsg := tgbotapi.NewEditMessageText(chatID, callbackQuery.Message.MessageID, "Предметы:")
	editMsg.ReplyMarkup = updateInlineKeyboard(callbackQuery.Message, itemID) // Функция для обновления клавиатуры
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Ошибка при обновлении сообщения: %v", err)
	}
}

func sellItems(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	itemID := ArrayItemShell[chatID]
	if itemID > 0 {
		count, iitems := a.PlayerItemCount(chatID)
		if count == 0 {
			item := a.PlayerItems(chatID)
			for _, i := range item {
				if i.ID == itemID {
					a.plusPlayerGold(int(chatID), i.Price)
					break
				}
			}
			a.DeleteItem(itemID, int(chatID)) // создать логику если надето на персоонаже то не удалять
			a.deleteItem(chatID, itemID)
			delete(ArrayItemShell, chatID)

		} else {
		itemBreak:
			for _, i := range iitems {
				if i.itemID == itemID {
					msg := tgbotapi.NewMessage(chatID, "Предмет не продан!")
					bot.Send(msg)
					break itemBreak
				}
			}
		}

		selectItem(bot, callbackQuery, 0)
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Ошибка при отправке ответа на callback: %v", err)
	}
}

func updateInlineKeyboard(message *tgbotapi.Message, itemID int) *tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	itemsp := a.PlayerItems(message.Chat.ID)
	for _, i := range itemsp {
		item := fmt.Sprintf("- %s (%s: +%d) 💰 - %d", i.Name, i.Stat, i.Value, i.Price)
		button := tgbotapi.NewInlineKeyboardButtonData(item, fmt.Sprintf("action_%d", i.ID))
		if itemID == i.ID { // Если элемент выбран, добавляем галочку
			button.Text = "✅ " + item
		}
		buttons = append(buttons, button)
	}

	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += 1 {
		inlineKeyboard = append(inlineKeyboard, []tgbotapi.InlineKeyboardButton{buttons[i]})
	}

	inlineKeyboard = append(inlineKeyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Продать", "item_sell"),
	})

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	var monster *Monster

	player := a.getPlayer(message.From.ID)
	if player == nil {
		// Начало игры, запрос имени и выбора класса
		startGame(bot, message)
		return
	}

	delEmojiMessage := parseEmoji(message.Text)

	step := countKillPlayer(message.Chat.ID, 0)

	monster = nextMonster(bot, message, step, player, delEmojiMessage)

	getclass := a.getClass(player.Class)
	i := ItemMap[message.Chat.ID]

	// Основные действия игрока
	switch delEmojiMessage {
	case "start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Добро пожаловать в игру!")
		bot.Send(msg)
		showMenu(bot, message.Chat.ID)
	case "Атака", getclass.Button.Name:
		// monster = nextMonster(bot, message, monster, step, player, delEmojiMessage)
		showMenuBattle(bot, message.Chat.ID)
		handleAttack(bot, player, monster, message.Chat.ID, message)
	case "Лечение x":
		if countPotion[message.Chat.ID] > 0 {
			pot := potion(message.Chat.ID, 1)
			if pot >= -1 {
				handleHeal(bot, player, message)
			}
		}
	case "Убежать", "Обойти":
		handleRun(bot, player, monster, message.Chat.ID, step, message.Text)
	case "help":
		class := getclass
		msg := tgbotapi.NewMessage(message.Chat.ID, class.Button.Help)
		bot.Send(msg)
	case "Подземелье":
		potion(message.Chat.ID, -1)
	case "Таверна":
		showTavern(bot, message.Chat.ID)
	case "Снять номер":
		showTavernRoom(bot, message)
	case "Уйти", "Назад":
		showMenu(bot, message.Chat.ID)
	case "Статистика":
		player.statisticPlayer(bot, message)
	case "Выйти из подземелья":
		delete(counts, message.Chat.ID)
		showMenu(bot, message.Chat.ID)
	case "Заплатить":
		if !checkGold(player.Gold, 50) {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Золотишек то нет!")
			bot.Send(msg)
			return
		}
		a.updatePlayerGold(int(message.Chat.ID), 50)
		handleHeal(bot, player, message)
		showTavern(bot, message.Chat.ID)
	case "Покупка":
		showShop(bot, message)
	case "Продажа":
		showShopPlayer(bot, message)
		// получение предмета
	case a.GetItem(message.Text).Name:
		showItem(bot, message)
	case "Купить":
		if !checkGold(player.Gold, i.Price) {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Золотишек то нет!")
			bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Преобретен предмет: %s", i.Name))
		bot.Send(msg)
		a.updatePlayerGold(int(message.Chat.ID), i.Price)
		a.AddItem(message.Chat.ID, i)
		delete(ItemMap, message.Chat.ID)
		showShop(bot, message)
	case "Инвентарь":
		showInventory(bot, message)
	case "Торговец":
		choiceShop(bot, message)
	}
}

// проверка суммы игрока
func checkGold(summ int, summT int) bool {
	return summ > summT
}

func choiceShop(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rand.NewSource(time.Second.Microseconds())
	say := []string{"Чего желаешь?", "Какие планы?", "Слушаю!", "Смотрю ты с барахлом!"}
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Покупка"),
			tgbotapi.NewKeyboardButton("Продажа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	sayR := rand.Intn(len(say))
	msg := tgbotapi.NewMessage(message.Chat.ID, say[sayR])
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// отображаем выбранный товар
func showItem(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	item := a.GetItem(message.Text)

	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Купить"),
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)

	str := "Предмет: %s\n%s - %d ~ %d\nЦена - %d"
	ItemMap[message.Chat.ID] = item
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(str, item.Name, item.Stat, item.Value, item.Value+RandVal[2], item.Price))
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// записывае рандомные числа для магазина
var ItemInt []int

func showShop(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	timeMinute := 30 * time.Minute
	items := a.GetItems()
	timeNow := time.Now()
	rand.NewSource(time.Now().Unix())

	dur := a.TimeDuration("shop")

	if dur.IsZero() {
		a.InsertTime("shop", timeMinute)

		for i := 4; i > 0; i-- {
			r := rand.Intn(len(items))
			ItemInt = append(ItemInt, r)
		}
	} else if timeNow.After(dur) {
		ItemInt = []int{} // чистим от прошлых значений
		for i := 4; i > 0; i-- {
			r := rand.Intn(len(items))
			ItemInt = append(ItemInt, r)
		}

		a.UpdateTime("shop", timeMinute)
	}

	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(items[ItemInt[0]].Name),
			tgbotapi.NewKeyboardButton(items[ItemInt[1]].Name),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(items[ItemInt[2]].Name),
			tgbotapi.NewKeyboardButton(items[ItemInt[3]].Name),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Уйти"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Вот мои товары!")
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func showShopPlayer(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	player := a.getPlayer(message.From.ID)
	if player == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ваши данные не найдены.")
		bot.Send(msg)
		return
	}

	var buttons []tgbotapi.InlineKeyboardButton

	itemsp := a.PlayerItems(message.From.ID)
	for _, i := range itemsp {
		ArrayItemShell[message.From.ID] = i.ID

		item := fmt.Sprintf("- %s (%s: +%d) 💰 - %d", i.Name, i.Stat, i.Value, i.Price)
		button := tgbotapi.NewInlineKeyboardButtonData(item, fmt.Sprintf("action_%d", i.ID))
		buttons = append(buttons, button)
	}

	if len(itemsp) < 1 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Предметов нет.")
		bot.Send(msg)
		return
	}

	// Создаём массив строк для кнопок
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(buttons); i += 1 { // Размещаем кнопки в столбик
		inlineKeyboard = append(inlineKeyboard, []tgbotapi.InlineKeyboardButton{buttons[i]})
	}

	// Добавляем кнопку "Продать"
	inlineKeyboard = append(inlineKeyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Продать", "action_sell"),
	})

	msg := tgbotapi.NewMessage(message.Chat.ID, "Предметы:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	bot.Send(msg)
}

func showTavernRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	text := "50 золотых! Берешь?"

	msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+text)
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Заплатить"),
			tgbotapi.NewKeyboardButton("🚪 Уйти"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func nextRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rand.NewSource(time.Now().UnixMicro())
	texts := []string{"💬Ну что, монстр, как тебе мой уровень мастерства?",
		"💬Еще один враг на моем пути! Неужели они не понимают, что я — герой?",
		"💬Похоже, этот монстр не читал мой гайд по победе!",
		"💬Снова победа! Как же скучно побеждать таких слабаков!",
		"💬Монстр, ты был великолепен... в своих мечтах!",
		"💬Я думал, будет сложнее. Может, в следующий раз выбери кого-то посильнее?",
		"💬Еще один монстр в списке моих жертв. У кого-то явно неудачный день!",
		"💬Этот монстр не знал, с кем связался. Теперь он знает!",
		"💬Победа! Не забудьте оставить отзыв о моем мастерстве!",
		"💬Монстр, ты был хорош, но, увы, я — лучше!",
		"💬Тень повержена, но страх остается.",
		"💬Каждая победа — это лишь шаг к новой тьме.",
		"💬Монстр мертв, но его крики еще звучат в моей голове.",
		"💬Смерть одного — это начало страха для других.",
		"💬Я победил, но цена была высока.",
		"💬Кровь на моих руках, и это лишь начало.",
		"💬Победа — это иллюзия, скрывающая настоящую тьму.",
		"💬Каждый враг, которого я убиваю, делает меня немного более бездушным.",
		"💬Монстр пал, но его тень навсегда останется со мной.",
		"💬Я победил, но в этом мире нет места для истинного триумфа."}
	r := rand.Intn(len(texts) + 20)
	var text string
	if r >= 20 {
		text = ""
	} else {
		text = texts[r]
	}

	// Генерируем случайное число: 0 или 1
	randomValue := rand.Intn(3)

	// Основные кнопки
	button := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Вперед"),
		tgbotapi.NewKeyboardButton("Статистика"),
	}

	// Если randomValue равно 1, добавляем кнопку выход из подъземелья
	if randomValue == 1 {
		button = append(button, tgbotapi.NewKeyboardButton("Выйти из подземелья"))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(button...),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func (player *Player) statisticPlayer(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Формируем сообщение с параметрами игрока
	text := fmt.Sprintf(
		"Игрок: %s\nКласс: %s\nHP: %d\\%d\nАтака: %d\nЗащита: %d\nЛовкость: %d\nЗолото: %d\nXP: %d\\%d\nУровень: %d\n",
		player.Name, player.Class, player.HP, player.MaxHP, player.Attack, player.Defense, player.Agility,
		player.Gold, player.XP, XPPerLevel*player.Level, player.Level,
	)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	bot.Send(msg)
}

func nextMonster(bot *tgbotapi.BotAPI, message *tgbotapi.Message, step int, player *Player, text string) *Monster {
	var monster *Monster
	if step >= 5 {
		monster = getMonsterBoss(player.Stage)
	} else {
		monster = getMonsterLite(player.Stage)
	}

	if text == "Атака" {
		return monster
	} else if text == "Вперед" || text == "Подземелье" {
		str := fmt.Sprintf("Вы встретили %s!\nВыберите действие:", monster.Name)
		msg := tgbotapi.NewMessage(message.Chat.ID, str)
		buttons := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Атака"),
				tgbotapi.NewKeyboardButton("Обойти"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Статистика"),
			),
		)
		msg.ReplyMarkup = buttons
		bot.Send(msg)
	}
	return monster
}

func startGame(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "👋 Привет! Выберите класс вашего персонажа:")
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛡 Воин"),
			tgbotapi.NewKeyboardButton("🧙‍♂️ Маг"),
			tgbotapi.NewKeyboardButton("🏹 Лучник"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Некромант"),
			tgbotapi.NewKeyboardButton("Жрец"),
			tgbotapi.NewKeyboardButton("Паладин"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)

	// парсим класс убирая эмодзи
	class := parseEmoji(message.Text)

	go func() {
		for {
			if message != nil && class == "Воин" || class == "Маг" ||
				class == "Лучник" || class == "Некромант" || class == "Жрец" || class == "Паладин" {
				a.addPlayer(message.From.ID, message.From.UserName, class)
				msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Вы выбрали класс: %s!", class))
				bot.Send(msg)
				potion(message.Chat.ID, -1)
				showMenu(bot, message.Chat.ID)
				break
			}
		}
	}()
}

func showMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🍻 Таверна"),
			tgbotapi.NewKeyboardButton("🔻 Подземелье"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧝‍♀️ Торговец"),
			tgbotapi.NewKeyboardButton("📋 Статистика"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎒 Инвентарь"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func showMenuBattle(bot *tgbotapi.BotAPI, chatID int64) {
	player := a.getPlayer(chatID)
	playerClass := a.getClass(player.Class)
	pot := potion(chatID, 0)

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚔️ Атака"),
			tgbotapi.NewKeyboardButton(fmt.Sprintf("💚 Лечение x%d", pot)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(playerClass.Button.NameSkill),
			tgbotapi.NewKeyboardButton("🏃‍♂️ Убежать"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

type stateFight struct {
	hpmonster int
	state     bool
}

func updateMonsterHP(monster *Monster, playerDamage int, chatID int64) *stateFight {
	if mon, ok := MonsterHP[chatID]; ok {
		// Ключ существует
		if !mon.state {
			mon.hpmonster = monster.HP
			mon.state = true
		} else {
			mon.hpmonster -= playerDamage
		}
	} else {
		// Ключа нет, создаем новый объект
		MonsterHP[chatID] = &stateFight{
			hpmonster: monster.HP,
			state:     true,
		}
	}

	return MonsterHP[chatID]
}

// из лута только бутылки с хилом
func lootMobs(chatID int64, lsBoss bool) string {
	rand.NewSource(time.Now().UnixMilli())
	r := rand.Intn(6)

	var str string
	switch r {
	case 1:
		potion(chatID, 5) // добавляем 1 potion
		str = fmt.Sprintln("Вы получили 1 зелье")
	case 2:
		// Выпадение случайного предмета
		items := a.GetItems()
		randomItem := items[rand.Intn(len(items))]

		randomItemNew := Item{
			ID:    randomItem.ID,
			Name:  randomItem.Name,
			Stat:  randomItem.Stat,
			Value: randomItem.Value,
			Price: randomItem.Price / 2,
		}

		a.AddItem(chatID, randomItemNew)

		str = fmt.Sprintf("Вы получили предмет: 📦 %s!\n", randomItemNew.Name)
	default:
		str = ""
	}
	return str
}

func handleAttack(bot *tgbotapi.BotAPI, player *Player, monster *Monster, chatID int64, message *tgbotapi.Message) {
	uph := updateMonsterHP(monster, 0, chatID)
	if player.Level%4 >= 0 {
		x := player.Level / 5
		if x == 0 {
			x = 1
		}
		monster = upgradeMonster(monster, Round(float64(x)))
	}

	class := a.getClass(player.Class)

	var ulimate *Ultimates
	if message.Text == class.Button.Name {
		ulimate = ultimate(class.Name)
	} else {
		ulimate = &Ultimates{
			Value: 0,
		}
	}

	if uph.state {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы атакуете %s!", monster.Name))
		bot.Send(msg)

		// Сражение
		// playerDamage, text := calculateDamage(&class, *monster, ulimate) // player.Attack - monster.Power/2
		playerDamage, text := calculateDamage(player, *monster, ulimate) // player.Attack - monster.Power/2
		if playerDamage < 0 {
			playerDamage = 0
		}

		if text != "" {
			text = "Вы нанесли ❗️%d критического урона. HP монстра: %d"
		} else {
			text = "Вы нанесли 🗡%d урона. HP монстра: %d"
		}

		uph := updateMonsterHP(monster, playerDamage, chatID) // Обновление HP монстра
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf(text, playerDamage, uph.hpmonster))
		bot.Send(msg)

		if uph.hpmonster > 0 {
			monsterDamage := monster.AtkPower - class.BaseDef/2 - ulimate.Value
			if monsterDamage < 0 {
				monsterDamage = 0
			}
			player.HP -= monsterDamage
			a.updatePlayerHP(player.ID, player.HP) // Сохранение нового HP игрока в БД
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("[%s] атакует! Вы потеряли %d HP. Ваши HP: %d",
				monster.Name, monsterDamage, player.HP))
			bot.Send(msg)
		} else {
			countKillPlayer(chatID, 1)
			player.XP += monster.XP
			player.Gold += monster.Gold
			player.Stage++

			levelUp(bot, player, chatID)
			a.updatePlayer(player)
			delete(MonsterHP, chatID) // удаляем прошлого монстра

			loot := lootMobs(chatID, monster.IsBoss)

			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы победили %s!\nЗолото: %d, Опыт: %d\n%s",
				monster.Name, player.Gold, player.XP, loot))
			bot.Send(msg)

			nextRoom(bot, message)
			// если босс и побежден отправляем в меню
			if monster.IsBoss {
				delete(counts, chatID)
				showMenu(bot, chatID)
			}
		}

		if player.HP <= 0 {
			msg := tgbotapi.NewMessage(chatID, "Вы погибли. Игра окончена.")
			bot.Send(msg)
			a.deletePlayer(player.ID) // после смерти удаляем героя
			startGame(bot, message)
			delete(counts, chatID) // после смерти удаляем подсчет монстров
			delete(countPotion, chatID)
		}
	}
}

func handleHeal(bot *tgbotapi.BotAPI, player *Player, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	if message.Text == "Заплатить" {
		a.updatePlayerHP(int(chatID), player.MaxHP)
		return
	}

	rand.NewSource(time.Now().UnixNano())
	heal := rand.Intn(14) + 26*player.Level

	player.HP += heal
	if player.HP > player.MaxHP { // проверяем чтобы HP не привышало максимума
		player.HP = player.MaxHP
	}

	a.updatePlayer(player)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы исцелились на %d. Ваши HP: %d/%d", heal, player.HP, player.MaxHP))
	bot.Send(msg)
	showMenuBattle(bot, chatID)
}

// Бегство от монстра, при провале получение урона.
func handleRun(bot *tgbotapi.BotAPI, player *Player, monster *Monster, chatID int64, step int, text string) {
	rand.NewSource(time.Now().UnixMicro())
	ran := rand.Intn(a.getClass(player.Class).Agility)

	if text == "Обойти" {
		if ran == 0 {
			text := []string{"Вы заметили, как " + monster.Name + " обернулся, и ваше сердце забилось быстрее.",
				"Попытка обойти " + monster.Name + " оказалась неудачной — оно мгновенно заметило вас.",
				"Внезапный шум привлек внимание " + monster.Name + ", и вас поймали врасплох.",
				"Вы не рассчитали расстояние и споткнулись, привлекая внимание " + monster.Name + ".",
				monster.Name + " резко повернулся, и вы поняли, что ваш план провалился."}
			r := rand.Intn(len(text))
			msg := tgbotapi.NewMessage(chatID, text[r])
			bot.Send(msg)
		} else {
			player.Stage++
			a.updatePlayer(player)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы успешно обошли %d монстра.", step))
			bot.Send(msg)
			step++
			return
		}
	}

	if step >= 5 {
		msg := tgbotapi.NewMessage(chatID, "❌ Убежать невозможно!")
		bot.Send(msg)
		return
	}

	if ran == 0 {
		monsterDamage := monster.AtkPower - player.Defense/2
		if monsterDamage < 0 {
			monsterDamage = 0
		}
		player.HP -= monsterDamage
		a.updatePlayerHP(player.ID, player.HP) // Сохранение нового HP игрока в БД
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("При попытке сбежать от %s вам нанесли урона %d HP. Ваши HP: %d", monster.Name, monsterDamage, player.HP))
		bot.Send(msg)
		showMenuBattle(bot, chatID)
	} else {
		player.Stage++
		a.updatePlayer(player)
		msg := tgbotapi.NewMessage(chatID, "Вы убежали от монстра.")
		bot.Send(msg)
		showMenuBattle(bot, chatID)
	}
}

// monster
func getMonsterLite(stage int) *Monster {
	monsters := []Monster{
		{"Гоблин", 40, 7, 5, 10, 3, false},
		{"Орк", 60, 10, 7, 8, 5, false},
		{"Мумия", 65, 13, 6, 6, 6, false},
		{"Вурдолак", 45, 6, 8, 7, 4, false},
		{"Скелет воин", 65, 9, 9, 8, 2, false},
		{"Демон", 70, 12, 8, 5, 4, false},
		{"Привидение", 60, 5, 4, 10, 7, false},
		{"Дракончик", 67, 14, 6, 6, 5, false},
		{"Леший", 65, 11, 7, 5, 6, false},
		{"Сирена", 55, 8, 5, 9, 6, false},
		{"Гигантский паук", 55, 10, 5, 7, 4, false},
		{"Зомби", 45, 6, 7, 6, 3, false},
		{"Фея", 40, 4, 3, 12, 8, false},
		{"Тень", 35, 7, 6, 11, 6, false},
		{"Водяной дух", 65, 9, 5, 8, 7, false},
		{"Суккуб", 55, 10, 4, 9, 7, false},
		// Name:     "",
		// HP:       0,
		// AtkPower: 0,
		// Def:      0,
		// XP:       XPPerLevel,
		// Gold:     0,
		// Boss: bool
	}

	rand.NewSource(time.Now().UnixNano())
	return &monsters[stage%len(monsters)]
}

// Boss
func getMonsterBoss(stage int) *Monster {
	monsters := []Monster{
		{"💀Дракон", 180, 16, 7, 20, 15, true},
		{"💀Дракула", 150, 17, 5, 17, 9, true},
		{"💀Троль", 190, 15, 6, 26, 8, true},
		{"💀Минотавр", 190, 15, 5, 28, 6, true},
		{"💀Ледяной гигант", 170, 11, 8, 3, 5, true},
		{"💀Костяной дракон", 175, 14, 7, 4, 5, true},
		// Name:     "",
		// HP:       0,
		// AtkPower: 0,
		// Def:      0,
		// XP:       XPPerLevel,
		// Gold:     0,
		// Boss: bool
	}

	rand.NewSource(time.Now().UnixNano())
	return &monsters[stage%len(monsters)]
}

// расчет повышения уровня
func levelUp(bot *tgbotapi.BotAPI, player *Player, chatID int64) {
	if player.XP >= XPPerLevel*player.Level {
		class := a.getClass(player.Class) // classes[player.Class]
		player.Level++
		player.MaxHP += 7 * player.Level
		player.Attack += class.BaseAtk / 2
		player.Defense += class.BaseDef / 2
		player.HP = player.MaxHP // устанавливаем базовое значение HP

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🌟Поздравляем! Вы достигли %d уровня!🌟", player.Level))
		bot.Send(msg)
	}
}

// подсчет пройденных монстров
func countKillPlayer(chaID int64, count int) int {
	counts[chaID] += count
	return counts[chaID]
}

func showTavern(bot *tgbotapi.BotAPI, chatID int64) {
	// say := []string{"Бодрость переполняет вас!"}

	say := []string{`Громкий скрип половиц прервал гул таверны, когда тяжёлый сапог ступил на порог. Взгляд каждого обитателя в мгновение ока устремился к фигуре в плаще. Шепот пронесся по залу, а трактирщик, зловеще прищурившись, убрал кружку с пивом, словно не замечая пришедшего. "Тебя здесь не ждали", — прорычал он.`,
		`Дверь с громким стуком распахнулась, и на пороге показался мужчина с усталым лицом. Несколько мгновений тишины, а затем кто-то из угла злобно бросил: "Мы думали, ты сгинул в горах". Подошедший трактирщик без слов указал на дверь, не давая и шанса заговорить. Взгляды за спинами говорили больше, чем слова.`,
		`В тени таверны огни казались теплыми, но как только странник пересёк порог, атмосфера будто замерла. Каждый, кто ещё мгновение назад смеялся или пил, отвёл взгляд. "Ты зря вернулся," — тихо, но с угрозой в голосе, сказал хозяин заведения, поднимая руку в предупреждающем жесте.`,
		`Тихий звон дверного колокольчика прозвучал в зале, когда фигура с капюшоном шагнула внутрь. Холодный ветер ворвался следом. Трактирщик остановил вытирание кружки и смерил незваного гостя взглядом. "Насытился бы уже скитаниями, Ренальд. Здесь для тебя места нет." Слова повисли в воздухе, как приговор.`,
		`Дверь таверны распахнулась с силой, и тень шагнула внутрь. Разговоры стихли. За стойкой трактирщик хмуро посмотрел на пришедшего, а в дальнем углу кто-то тихо выругался. "Опять ты... Думаешь, что кто-то рад тебя здесь видеть?" — произнёс голос из толпы, вызывая одобрительный ропот.`}
	r := rand.Intn(len(say))
	msg := tgbotapi.NewMessage(chatID, say[r])
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Снять номер"),
			tgbotapi.NewKeyboardButton("🚪 Уйти"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// убираме из строки эмоджи
func parseEmoji(text string) string {
	re := regexp.MustCompile("[А-Яа-яA-Za-z.*$]+")
	texts := re.FindAllString(text, -1)
	i := len(texts) - 1
	if i >= 1 {
		if texts[0] != "" && texts[1] != "" {
			text = strings.Join(texts, " ")
		}
	} else {
		text = re.FindString(text)
	}
	return text // re.FindString(text)
}

// прибавляем 1 потион до нужного количества
func potion(chatID int64, count int) int {
	// добавляем 3 потиона
	if count == -1 {
		delete(countPotion, chatID)
		countPotion[chatID] += 2
	}

	if count == 5 && countPotion[chatID] < 2 {
		countPotion[chatID] += 1
		return countPotion[chatID]
	}

	// Проверяем, если количество зелья для chatID равно 0
	if countPotion[chatID] <= 0 {
		return countPotion[chatID] // Возвращаем текущее значение (0 или меньше)
	}

	if countPotion[chatID] > 0 && count != 5 {
		countPotion[chatID] -= count
	}

	return countPotion[chatID]
}

// Калькуляция урона с критом
func calculateDamage(attacker *Player, target Monster, ulimate *Ultimates) (int, string) {
	var text string
	// Инициализируем генератор случайных чисел для критического удара
	rand.NewSource(time.Now().UnixNano())

	// Базовый урон
	// baseDamage := attacker.BaseAtk + ulimate.Value - target.Def
	baseDamage := attacker.Attack + ulimate.Value - target.Def

	// Учитываем, что урон не может быть меньше 0
	if baseDamage < 0 {
		baseDamage = 0
	}

	// Критический удар
	critChance := 0.1 // 10% шанс на крит

	playAgil := float64(attacker.Agility) / float64(100)
	fmt.Println(playAgil, critChance, critChance+playAgil)
	isCrit := rand.Float64() <= critChance+playAgil

	// Множитель крита
	critMultiplier := 2.0
	if isCrit {
		text = "Критический удар!"
		baseDamage = int(float64(baseDamage) * critMultiplier)
	}

	// Рандомный фактор (например, урон варьируется в пределах ±10%)
	randomFactor := 0.9 + rand.Float64()*0.2 // диапазон от 0.9 до 1.1
	finalDamage := int(float64(baseDamage) * randomFactor)

	return finalDamage, text
}

// увеличения сложности мобов
func upgradeMonster(monster *Monster, up int) *Monster {
	m := &Monster{
		Name:     monster.Name,
		HP:       monster.HP * up,
		AtkPower: monster.AtkPower * up,
		Def:      monster.Def - 1*up,
		XP:       monster.XP * up,
		Gold:     monster.Gold * up,
		IsBoss:   monster.IsBoss,
	}

	return m
}

func Round(x float64) int {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.1 {
		return int(t + math.Copysign(1, x))
	}
	return int(t)
}
