package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

type Player struct {
	ID      int
	Name    string
	Class   string
	HP      int
	Attack  int
	Defense int
	Gold    int
	Stage   int
	XP      int
	Level   int
	Count   int
}

type Items struct {
	ID    int
	Name  string
	Param int
	Count int
}

type Monster struct {
	Name     string
	HP       int
	AtkPower int
	Def      int
	XP       int
	Gold     int
	Boss     bool
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
	Name  string
	Help  string
	Value int
}

var classes = map[string]Class{
	"Воин": {
		Name:    "Воин",
		BaseHP:  150,
		BaseAtk: 20,
		BaseDef: 10,
		Agility: 4,
		Button: Skils{
			Name:  "Оборона",
			Help:  "Способность война: Оборона\nВоин прикрывается щитом и получает 0 урона.",
			Value: 100,
		},
	},
	"Маг": {
		Name:    "Маг",
		BaseHP:  100,
		BaseAtk: 30,
		BaseDef: 5,
		Agility: 3,
		Button: Skils{
			Name:  "Ледяная глыба",
			Help:  "Способность маг: Ледяная глыба\nМаг превращается в глыбу и при этом не получает урон",
			Value: 100,
		},
	},
	"Лучник": {
		Name:    "Лучник",
		BaseHP:  120,
		BaseAtk: 25,
		BaseDef: 7,
		Agility: 4,
		Button: Skils{
			Name:  "Отскок",
			Help:  "Способность лучника: Отскок\nЛучник отпрыгивает от врага тем самым пропуская его",
			Value: 100,
		},
	},
}

const XPPerLevel = 180

var counts = make(map[int64]int)
var countPotion = make(map[int64]int)

func main() {
	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI("7524529714:AAHWPV-x44cN9BWIa9JAq_xpyl3Uqbl4dIY")
	if err != nil {
		log.Panic("ошибка подключения: ", err)
	}

	bot.Debug = false

	// Инициализация БД
	db, err := sql.Open("sqlite3", "./files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание таблиц
	createTables(db)

	// Обработка сообщений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, db, update.Message)
		}
	}
}

func createTables(db *sql.DB) {
	querys := []string{`CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		class TEXT,
		hp INTEGER,
		attack INTEGER,
		defense INTEGER,
		gold INTEGER,
		stage INTEGER,
		xp INTEGER,
		level INTEGER
	);`, `CREATE TABLE IF NOT EXISTS items (
		id INTEGER,
		name TEXT,
		param INTEGER,
		count INTEGER
	);`}
	for _, query := range querys {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	var monster *Monster

	player := getPlayer(db, message.From.ID)
	if player == nil {
		// Начало игры, запрос имени и выбора класса
		startGame(bot, db, message)
		return
	}

	delEmojiMessage := parseEmoji(message.Text)

	step := countKillPlayer(message.Chat.ID, 0)

	monster = nextMonster(bot, message, step, player, delEmojiMessage)

	// Основные действия игрока
	switch delEmojiMessage {
	case "start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Добро пожаловать в игру!")
		bot.Send(msg)
		showMenu(bot, message.Chat.ID)
	case "Атака":
		// monster = nextMonster(bot, message, monster, step, player, delEmojiMessage)
		showMenuBattle(bot, db, message.Chat.ID)
		handleAttack(bot, db, player, monster, message.Chat.ID, message)
	case "Лечение":
		if countPotion[message.Chat.ID] > 0 {
			pot := potion(message.Chat.ID, 1)
			if pot >= -1 {
				handleHeal(bot, db, player, message.Chat.ID)
			}
		}
	case "Убежать", "Обойти":
		handleRun(bot, db, player, monster, message.Chat.ID, step, message.Text)
	case "help":
		player := getPlayer(db, message.Chat.ID)
		class := classes[player.Class]
		msg := tgbotapi.NewMessage(message.Chat.ID, class.Button.Help)
		bot.Send(msg)
	case "Подземелье":
		potion(message.Chat.ID, -1)
	case "Таверна":
		showTavern(bot, message.Chat.ID)
	case "Уйти":
		showMenu(bot, message.Chat.ID)
	case classes[player.Class].Button.Name:

	case "Статистика":
		statisticPlayer(bot, message, player)
	case "Выйти":
		showMenu(bot, message.Chat.ID)
	}
}

func nextRoom(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rand.NewSource(time.Now().UnixMicro())
	texts := []string{"Ну что, монстр, как тебе мой уровень мастерства?",
		"Еще один враг на моем пути! Неужели они не понимают, что я — герой?",
		"Похоже, этот монстр не читал мой гайд по победе!",
		"Снова победа! Как же скучно побеждать таких слабаков!",
		"Монстр, ты был великолепен... в своих мечтах!",
		"Я думал, будет сложнее. Может, в следующий раз выбери кого-то посильнее?",
		"Еще один монстр в списке моих жертв. У кого-то явно неудачный день!",
		"Этот монстр не знал, с кем связался. Теперь он знает!",
		"Победа! Не забудьте оставить отзыв о моем мастерстве!",
		"Монстр, ты был хорош, но, увы, я — лучше!",
		"Тень повержена, но страх остается.",
		"Каждая победа — это лишь шаг к новой тьме.",
		"Монстр мертв, но его крики еще звучат в моей голове.",
		"Смерть одного — это начало страха для других.",
		"Я победил, но цена была высока.",
		"Кровь на моих руках, и это лишь начало.",
		"Победа — это иллюзия, скрывающая настоящую тьму.",
		"Каждый враг, которого я убиваю, делает меня немного более бездушным.",
		"Монстр пал, но его тень навсегда останется со мной.",
		"Я победил, но в этом мире нет места для истинного триумфа."}
	r := rand.Intn(len(texts))

	// Генерируем случайное число: 0 или 1
	randomValue := rand.Intn(3)

	// Основные кнопки
	button := []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton("Вперед"),
		tgbotapi.NewKeyboardButton("Статистика"),
	}

	// Если randomValue равно 1, добавляем случайную кнопку
	if randomValue == 1 {
		button = append(button, tgbotapi.NewKeyboardButton("Выйти из подземелья"))
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+texts[r])
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(button...),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func statisticPlayer(bot *tgbotapi.BotAPI, message *tgbotapi.Message, player *Player) {
	// Формируем сообщение с параметрами игрока
	text := fmt.Sprintf(
		"Игрок: %s Класс: %s\nHP: %d\nАтака: %d\nЗащита: %d\nЗолото: %d\nXP: %d\nУровень: %d\n",
		player.Name, player.Class, player.HP, player.Attack, player.Defense,
		player.Gold, player.XP, player.Level,
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

func startGame(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "👋 Привет! Выберите класс вашего персонажа:")
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🛡 Воин"),
			tgbotapi.NewKeyboardButton("🧙‍♂️ Маг"),
			tgbotapi.NewKeyboardButton("🏹 Лучник"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)

	// парсим класс убирая эмодзи
	class := parseEmoji(message.Text)

	go func() {
	stop:
		for {
			if message != nil && (class == "Воин" || class == "Маг" || class == "Лучник") {
				addPlayer(db, message.From.ID, message.From.UserName, class)
				msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Вы выбрали класс: %s!", class))
				bot.Send(msg)
				potion(message.Chat.ID, -1)
				showMenu(bot, message.Chat.ID)
				break stop
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
			tgbotapi.NewKeyboardButton("Статистика"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func showMenuBattle(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64) {
	player := getPlayer(db, chatID)
	playerClass := classes[player.Class]
	pot := potion(chatID, 0)

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚔️ Атака"),
			tgbotapi.NewKeyboardButton(fmt.Sprintf("💚 Лечение x%d", pot)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(playerClass.Button.Name),
			tgbotapi.NewKeyboardButton("🏃‍♂️ Убежать"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

func addPlayer(db *sql.DB, userID int64, name, class string) {
	playerClass := classes[class]
	query := `INSERT INTO players (id, name, class, hp, attack, defense, gold, stage, xp, level) VALUES (?, ?, ?, ?, ?, ?, 0, 1, 0, 1)`
	_, err := db.Exec(query, userID, name, class, playerClass.BaseHP, playerClass.BaseAtk, playerClass.BaseDef)
	if err != nil {
		log.Fatal(err)
	}
}

func getPlayer(db *sql.DB, userID int64) *Player {
	query := `SELECT id, name, class, hp, attack, defense, gold, stage, xp, level FROM players WHERE id = ?`
	row := db.QueryRow(query, userID)

	var player Player
	err := row.Scan(&player.ID, &player.Name, &player.Class, &player.HP, &player.Attack, &player.Defense,
		&player.Gold, &player.Stage, &player.XP, &player.Level)
	if err != nil {
		return nil
	}
	return &player
}

func updatePlayer(db *sql.DB, player *Player) {
	query := `UPDATE players SET hp = ?, attack = ?, defense = ?, gold = ?, stage = ?, xp = ?, level = ? WHERE id = ?`
	_, err := db.Exec(query, player.HP, player.Attack, player.Defense,
		player.Gold, player.Stage, player.XP, player.Level, player.ID)
	if err != nil {
		log.Fatal(err)
	}
}

func updatePlayerHP(db *sql.DB, playerID int, newHP int) error {
	query := `UPDATE players SET hp = ? WHERE id = ?`
	_, err := db.Exec(query, newHP, playerID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении HP игрока: %v", err)
	}
	return nil
}

type stateFight struct {
	hpmonster int
	state     bool
}

var MonsterHP = make(map[int64]*stateFight)

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

func handleAttack(bot *tgbotapi.BotAPI, db *sql.DB, player *Player, monster *Monster, chatID int64, message *tgbotapi.Message) {
	uph := updateMonsterHP(monster, 0, chatID)
	class := classes[player.Class]

	if uph.state {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы атакуете %s!", monster.Name))
		bot.Send(msg)

		// Сражение
		playerDamage, text := calculateDamage(class, *monster) // player.Attack - monster.Power/2
		if playerDamage < 0 {
			playerDamage = 0
		}

		if text != "" {
			text = "Вы нанесли ❗️%d критического урона. HP монстра: %d"
		} else {
			text = "Вы нанесли %d урона. HP монстра: %d"
		}

		uph := updateMonsterHP(monster, playerDamage, chatID) // Обновление HP монстра
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf(text, playerDamage, uph.hpmonster))
		bot.Send(msg)

		if uph.hpmonster > 0 {
			monsterDamage := monster.AtkPower - class.BaseDef/2
			if monsterDamage < 0 {
				monsterDamage = 0
			}
			player.HP -= monsterDamage
			updatePlayerHP(db, player.ID, player.HP) // Сохранение нового HP игрока в БД
			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("%s атакует! Вы потеряли %d HP. Ваши HP: %d", monster.Name, monsterDamage, player.HP))
			bot.Send(msg)
		} else {
			countKillPlayer(chatID, 1)
			player.XP += 9 + monster.XP
			player.Gold += 2 + monster.Gold
			player.Stage++

			levelUp(bot, player, chatID)
			updatePlayer(db, player)
			delete(MonsterHP, chatID) // удаляем прошлого монстра

			msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы победили %s!\nЗолото: %d, Опыт: %d",
				monster.Name, player.Gold, player.XP))
			bot.Send(msg)

			nextRoom(bot, message)
			// если босс и побеждем отправляем в меню
			if monster.Boss {
				showMenu(bot, chatID)
			}
		}

		if player.HP <= 0 {
			msg := tgbotapi.NewMessage(chatID, "Вы погибли. Игра окончена.")
			bot.Send(msg)
			deletePlayer(db, player.ID) // после смерти удаляем героя
			startGame(bot, db, message)
			delete(counts, chatID) // после смерти удаляем подсчет монстров
			delete(countPotion, chatID)
		}
	}
}

func handleHeal(bot *tgbotapi.BotAPI, db *sql.DB, player *Player, chatID int64) {
	rand.NewSource(time.Now().UnixNano())
	player.HP += rand.Intn(30) + 36*player.Level/5
	updatePlayer(db, player)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы исцелились. Ваши HP: %d", player.HP))
	bot.Send(msg)
	showMenuBattle(bot, db, chatID)
}

// Бегство от монстра, при провале получение урона.
func handleRun(bot *tgbotapi.BotAPI, db *sql.DB, player *Player, monster *Monster, chatID int64, step int, text string) {
	rand.NewSource(time.Now().UnixMicro())
	ran := rand.Intn(classes[player.Class].Agility)

	if text == "Обойти" {
		if ran == 0 {
			text := []string{"Вы заметил, как " + monster.Name + " обернулся, и ваше сердце забилось быстрее.",
				"Попытка обойти " + monster.Name + " оказалась неудачной — оно мгновенно заметило вас.",
				"Внезапный шум привлек внимание " + monster.Name + ", и вас поймали врасплох.",
				"Вы не рассчитали расстояние и споткнулись, привлекая внимание " + monster.Name + ".",
				monster.Name + " резко повернулся, и вы поняли, что ваш план провалился."}
			r := rand.Intn(len(text))
			msg := tgbotapi.NewMessage(chatID, text[r])
			bot.Send(msg)
		} else {
			player.Stage++
			updatePlayer(db, player)
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
		updatePlayerHP(db, player.ID, player.HP) // Сохранение нового HP игрока в БД
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("При попытке сбежать от %s вам нанесли урона %d HP. Ваши HP: %d", monster.Name, monsterDamage, player.HP))
		bot.Send(msg)
		showMenuBattle(bot, db, chatID)
	} else {
		player.Stage++
		updatePlayer(db, player)
		msg := tgbotapi.NewMessage(chatID, "Вы убежали от монстра.")
		bot.Send(msg)
		showMenuBattle(bot, db, chatID)
	}
}

// удаление
func deletePlayer(db *sql.DB, playerID int) {
	query := `DELETE FROM players WHERE id = ?`
	_, err := db.Exec(query, playerID)
	if err != nil {
		log.Fatal(err)
	}
}

// monster
func getMonsterLite(stage int) *Monster {
	monsters := []Monster{
		{"Гоблин", 50, 7, 5, 10, 3, false},
		{"Орк", 70, 10, 7, 8, 5, false},
		{"Мумия", 75, 13, 6, 6, 6, false},
		{"🧟‍♀️ Вурдолак", 45, 6, 8, 7, 4, false},
		{"💀 Скелет воин", 65, 9, 9, 8, 2, false},
		{"Демон", 80, 12, 8, 5, 4, false},
		{"Привидение", 60, 5, 4, 10, 7, false},
		{"Дракончик", 70, 14, 6, 6, 5, false},
		{"Леший", 65, 11, 7, 5, 6, false},
		{"Сирена", 55, 8, 5, 9, 6, false},
		{"Гигантский паук", 75, 10, 5, 7, 4, false},
		{"Зомби", 50, 6, 7, 6, 3, false},
		{"Фея", 40, 4, 3, 12, 8, false},
		{"Тень", 65, 7, 6, 11, 6, false},
		{"Водяной дух", 70, 9, 5, 8, 7, false},
		{"Суккуб", 60, 10, 4, 9, 7, false},
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
		{"Дракон", 210, 16, 8, 20, 15, true},
		{"Дракула", 180, 17, 5, 17, 9, true},
		{"Троль", 210, 15, 7, 26, 8, true},
		{"Минотавр", 200, 15, 6, 28, 6, true},
		{"Ледяной гигант", 180, 11, 12, 3, 5, true},
		{"Костяной дракон", 185, 14, 10, 4, 5, true},
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

func levelUp(bot *tgbotapi.BotAPI, player *Player, chatID int64) {
	if player.XP >= XPPerLevel*player.Level {
		class := classes[player.Class]
		player.Level++
		player.HP += 7 * player.Level
		player.Attack += class.BaseAtk / 2
		player.Defense += class.BaseDef / 2

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
	// tgbotapi.NewSticker(chatID, tgbotapi.FileID(""))
	say := []string{`Громкий скрип половиц прервал гул таверны, когда тяжёлый сапог ступил на порог. Взгляд каждого обитателя в мгновение ока устремился к фигуре в плаще. Шепот пронесся по залу, а трактирщик, зловеще прищурившись, убрал кружку с пивом, словно не замечая пришедшего. "Тебя здесь не ждали", — прорычал он.`,
		`Дверь с громким стуком распахнулась, и на пороге показался мужчина с усталым лицом. Несколько мгновений тишины, а затем кто-то из угла злобно бросил: "Мы думали, ты сгинул в горах". Подошедший трактирщик без слов указал на дверь, не давая и шанса заговорить. Взгляды за спинами говорили больше, чем слова.`,
		`В тени таверны огни казались теплыми, но как только странник пересёк порог, атмосфера будто замерла. Каждый, кто ещё мгновение назад смеялся или пил, отвёл взгляд. "Ты зря вернулся," — тихо, но с угрозой в голосе, сказал хозяин заведения, поднимая руку в предупреждающем жесте.`,
		`Тихий звон дверного колокольчика прозвучал в зале, когда фигура с капюшоном шагнула внутрь. Холодный ветер ворвался следом. Трактирщик остановил вытирание кружки и смерил незваного гостя взглядом. "Насытился бы уже скитаниями, Ренальд. Здесь для тебя места нет." Слова повисли в воздухе, как приговор.`,
		`Дверь таверны распахнулась с силой, и тень шагнула внутрь. Разговоры стихли. За стойкой трактирщик хмуро посмотрел на пришедшего, а в дальнем углу кто-то тихо выругался. "Опять ты... Думаешь, что кто-то рад тебя здесь видеть?" — произнёс голос из толпы, вызывая одобрительный ропот.`}
	r := rand.Intn(len(say))
	msg := tgbotapi.NewMessage(chatID, say[r])
	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🚪 Уйти"),
		),
	)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// убираме из строки эмоджи
func parseEmoji(text string) string {
	re := regexp.MustCompile("[А-Яа-яA-Za-z.*$]+")
	return re.FindString(text)
}

// прибавляем 1 потион до нужного количества
func potion(chatID int64, count int) int {
	// cou := countPotion[chatID]
	// добавляем 3 потиона
	if count == -1 {
		delete(countPotion, chatID)
		countPotion[chatID] += 2
	}
	// Проверяем, если количество зелья для chatID равно 0
	if countPotion[chatID] <= 0 {
		return countPotion[chatID] // Возвращаем текущее значение (0 или меньше)
	}

	if countPotion[chatID] > 0 {
		countPotion[chatID] -= count
	}

	return countPotion[chatID]
}

// Калькуляция урона с критом
func calculateDamage(attacker Class, target Monster) (int, string) {
	var text string
	// Инициализируем генератор случайных чисел для критического удара
	rand.NewSource(time.Now().UnixNano())

	// Базовый урон
	baseDamage := attacker.BaseAtk - target.Def

	// Учитываем, что урон не может быть меньше 0
	if baseDamage < 0 {
		baseDamage = 0
	}

	// Критический удар (пример: шанс крита = 20%)
	critChance := 0.1 // 20% шанс на крит
	isCrit := rand.Float64() <= critChance

	// Множитель крита (обычно это 2x, но может быть другим)
	critMultiplier := 2.0
	if isCrit {
		text = fmt.Sprintln("Критический удар!")
		baseDamage = int(float64(baseDamage) * critMultiplier)
	}

	// Рандомный фактор (например, урон варьируется в пределах ±10%)
	randomFactor := 0.9 + rand.Float64()*0.2 // диапазон от 0.9 до 1.1
	finalDamage := int(float64(baseDamage) * randomFactor)

	return finalDamage, text
}

func ultimate(chatID int64, player *Player) {

}
