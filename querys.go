package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

type App struct {
	DB       *sql.DB
	Monsters []*Monster
}

const database = "files.db"

// значения для рандомных характеристик. используется в randomValue и showItem
var RandVal = []int{1, 3, 5}

func (a *App) ConnDB() {
	db, err := sql.Open("sqlite", "./"+database)
	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
}

// создание БД если нет
func CraeteDataBase() {
	file, err := os.ReadDir("./")
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range file {
		if f.Name() == database {
			return
		}
	}
	// Подготовка команды sqlite3 files.db
	cmd := exec.Command("sqlite3", database)

	// Запуск команды
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// создание таблиц
func (a *App) createTablesAndDB() {
	querys := []string{`CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		class TEXT,
		hp INTEGER,
		attack INTEGER,
		defense INTEGER,
		agility INTEGER,
		gold INTEGER,
		stage TEXT,
		xp INTEGER,
		level INTEGER,
		maxhp INTEGER
	);`, `CREATE TABLE IF NOT EXISTS items (
	id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
    name TEXT,
    stat TEXT,
    value INTEGER,
    price INTEGER,
    UNIQUE ("name") ON CONFLICT IGNORE
);
`, `CREATE TABLE IF NOT EXISTS quests (
		name     TEXT,
		body     TEXT,
		progress BOOLEAN,
		ID       INTEGER
		);`, `INSERT INTO "items" (name, stat, value, price) VALUES
("Кольцо Древнего Леса", "MAXHP", 15, 120),
("Меч Огненной Стужи", "Attack", 15, 300),
("Щит Светлого Воина", "Defense", 12, 180),
("Перчатки Ловкости", "Attack", 5, 80),
("Плащ Призрачной Тени", "Defense", 8, 100),
("Серебряный Амулет", "MAXHP", 12, 90),
("Книга Заклинаний", "Attack", 8, 140),
("Сапоги Ускользания", "Defense", 4, 50),
("Талисман Мудрости", "MAXHP", 10, 70),
("Мантия Ледяного Ветра", "Defense", 13, 200),
("Кристалл Души", "MAXHP", 25, 250),
("Кинжал Ядовитой Змеи", "Attack", 10, 180),
("Шлем Легендарного Героя", "Defense", 15, 250),
("Пояс Силы", "Attack", 12, 220),
("Эльфийский Лук", "Attack", 15, 320),
("Сумка Бесконечности", "MAXHP", 8, 60),
("Чаша Мудрости", "Defense", 8, 110),
("Медальон Силы", "MAXHP", 5, 40),
("Перья Летучей Мыши", "Attack", 3, 40),
("Ботинки Стремительного Ветра", "agility", 5, 350),
("Коготь Гепарда", "agility", 4, 250),
("Пояс Проворства", "agility", 3, 170),
("Перстень Ловкача", "agility", 2, 90),
("Шнурки Скорости", "agility", 1, 40),
("Поножи Ястреба", "agility", 4, 260),
("Накидка Хитрого Лиса", "agility", 3, 160),
("Браслет Рефлексов", "agility", 2, 100),
("Амулет Гибкости", "agility", 1, 45),
("Кольцо Молниеносности", "agility", 5, 380),
("Перчатки Фехтовальщика", "agility", 3, 180);`,
		`CREATE TABLE IF NOT EXISTS chanTicker (
    name TEXT,
    duration DATETIME DEFAULT CURRENT_TIMESTAMP
);`, `CREATE TABLE IF NOT EXISTS player_item (
		idItem INTEGER PRIMARY KEY AUTOINCREMENT,
		id INTEGER,
		name TEXT,
		stat TEXT,
		value INTEGER,
		price INTEGER
	);`, `CREATE TABLE quests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    reward TEXT NOT NULL,
    status TEXT NOT NULL,
    required_items TEXT
);`, `CREATE TABLE player_item_equip (
    equip_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id   INTEGER NOT NULL,
    item_id     INTEGER NOT NULL,
    slot        INTEGER NOT NULL, -- 1 или 2 (слот экипировки)
    equipped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_player FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
    CONSTRAINT uq_player_slot UNIQUE (player_id, slot) -- один предмет в один слот
);`,
		`CREATE TABLE IF NOT EXISTS monsters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    hp INTEGER,
    atk_power INTEGER,
    def INTEGER,
    xp INTEGER,
    gold INTEGER,
    is_boss BOOLEAN
);`,
		`INSERT INTO monsters (name, hp, atk_power, def, xp, gold, is_boss) VALUES 
('Кобольд-разведчик', 35, 6, 2, 15, 10, 0),
('Гоблин', 45, 8, 4, 20, 15, 0),
('Фея', 40, 10, 2, 25, 20, 0),
('Зомби', 60, 10, 5, 25, 12, 0),
('Вурдолак', 65, 12, 6, 30, 20, 0),
('Скелет воин', 75, 11, 9, 35, 25, 0),
('Орк', 90, 14, 8, 45, 35, 0),
('Мумия', 85, 15, 7, 50, 40, 0),
('Привидение', 55, 18, 3, 55, 45, 0),
('Сирена', 70, 16, 5, 60, 50, 0),
('Гигантский паук', 80, 17, 6, 65, 55, 0),
('Тень', 60, 22, 4, 70, 60, 0),
('Дракончик', 95, 19, 10, 80, 75, 0),
('Леший', 110, 16, 12, 85, 70, 0),
('Суккуб', 90, 21, 8, 90, 85, 0),
('Гарпия', 85, 20, 7, 95, 90, 0),
('Водяной дух', 75, 14, 10, 60, 50, 0),
('Темный культист', 95, 20, 8, 100, 80, 0),
('Минотавр-рекрут', 120, 22, 12, 110, 90, 0),
('Ледяной голем', 140, 18, 18, 120, 100, 0),
('Огненный элементаль', 100, 26, 6, 115, 95, 0),
('Василиск', 110, 21, 14, 130, 110, 0),
('Мимик', 80, 25, 15, 150, 500, 0),
('Химера', 130, 24, 13, 160, 140, 0),
('💀Троль', 400, 35, 20, 800, 600, 1),
('💀Дракула', 450, 45, 15, 1000, 800, 1),
('💀Дракон', 600, 50, 30, 1500, 1200, 1),
('💀Ледяной гигант', 580, 42, 35, 1600, 1100, 1),
('💀Костяной дракон', 550, 55, 25, 1650, 1150, 1),
('💀Минотавр', 650, 48, 28, 1700, 1300, 1),
('💀Медуза Горгона', 420, 46, 20, 1800, 1400, 1),
('💀Черная Вдова', 400, 52, 18, 1900, 1450, 1),
('💀Анубис', 520, 48, 30, 2000, 1500, 1),
('💀Химера-переросток', 500, 54, 25, 2100, 1600, 1),
('💀Железный Голем', 850, 40, 55, 2500, 2000, 1),
('💀Верховный Лич', 480, 65, 20, 3000, 2500, 1),
('💀Огненный Феникс', 450, 70, 15, 3200, 2800, 1),
('💀Хранитель леса', 750, 50, 40, 3500, 3000, 1),
('💀Кракен', 900, 60, 35, 4000, 3500, 1),
('💀Падший Ангел', 700, 75, 30, 4500, 4000, 1),
('💀Теневой Жнец', 550, 85, 25, 4800, 4500, 1),
('💀Повелитель Бездны', 1200, 90, 45, 10000, 10000, 1);`}
	for _, query := range querys {
		_, err := a.DB.Exec(query)
		if err != nil {
			fmt.Println("func createTablesAndDB - querys:", err.Error())
		}
	}

	createClass := []string{`-- Таблица для хранения классов
CREATE TABLE IF NOT EXISTS classes (
    name TEXT,
    base_hp INTEGER,
    base_atk INTEGER,
    base_def INTEGER,
    agility INTEGER
);`, `CREATE TABLE IF NOT EXISTS skills (
    name TEXT,
	name_skill TEXT,
    help TEXT,
    value_base_hp INTEGER,
    value_base_atk INTEGER,
    value_base_def INTEGER,
    value_agility INTEGER
);`, `INSERT INTO classes (name, base_hp, base_atk, base_def, agility) VALUES 
("Воин",160,18,12,3),
("Маг",90,28,4,5),
("Лучник",110,22,7,8),
("Некромант",105,25,6,4),
("Жрец",120,15,10,4),
("Паладин",150,20,15,2),
("Тёмный Рыцарь",140,24,10,3);`,
		`INSERT INTO skills (name, name_skill, help, value_base_hp, value_base_atk, value_base_def, value_agility) VALUES
("Воин", "Оборона", "Способность война: Оборона\nВоин прикрывается щитом и получает 0 урона.", 0, 0, 0, 0),
("Маг", "Ледяная глыба", "Способность маг: Ледяная глыба\nМаг превращается в глыбу и при этом не получает урон.", 0, 0, 0, 0),
("Лучник", "Отскок", "Способность лучника: Отскок\nЛучник отпрыгивает от врага тем самым пропуская его", 0, 0, 0, 0),
("Некромант", "Призыв мертвецов", "Способность некроманта: Призыв мертвецов\nНекромант призывает мертвецов, которые отвлекают врагов, снижая урон на 50% на один ход.", 0, 0, 0, 0),
("Жрец", "Исцеление", "Способность жреца: Исцеление\nЖрец использует магию, чтобы восстановить себе или союзнику 30 очков здоровья.", 0, 0, 0, 0),
("Паладин", "Священный щит", "Способность паладина: Священный щит\nПаладин окружает себя и союзников щитом, который снижает получаемый урон на 50% на один ход.", 0, 0, 0, 0),
("Тёмный Рыцарь", "Поглощение Жизни", "Способность тёмного рыцаря: Поглощение Жизни\nТёмный рыцарь наносит урон врагу и восстанавливает себе 20 очков здоровья.", 0, 0, 0, 0);
`,
	}

	var count int
	_ = a.DB.QueryRow(`select count(*) from classes`).Scan(&count)
	if count <= 0 {
		for _, query := range createClass {
			_, err := a.DB.Exec(query)
			if err != nil {
				log.Fatalln("func createTablesAndDB - createClass:", err)
			}
		}
	}
}

func (a *App) addPlayer(userID int64, name, class string) {
	playerClass := a.getClass(class)
	query := `INSERT INTO players (id, name, class, hp, attack, defense, agility, gold, stage, xp, level, maxhp) VALUES (?, ?, ?, ?, ?, ?, ?, 0, "home", 0, 1, ?)`
	_, err := a.DB.Exec(query, userID, name, class, playerClass.BaseHP, playerClass.BaseAtk, playerClass.BaseDef, playerClass.Agility, playerClass.BaseHP)
	if err != nil {
		log.Fatal("addPlayer: ", err)
	}
}

func (a *App) getPlayer(userID int64) *Player {
	query := `SELECT id, name, class, hp, attack, defense, agility, gold, stage, xp, level, maxhp FROM players WHERE id = ?`
	row := a.DB.QueryRow(query, userID)

	var player Player
	err := row.Scan(&player.ID, &player.Name, &player.Class, &player.HP, &player.Attack, &player.Defense, &player.Agility,
		&player.Gold, &player.Stage, &player.XP, &player.Level, &player.MaxHP)
	if err != nil {
		return nil
	}
	return &player
}

func (a *App) updatePlayer(player *Player) {
	query := `UPDATE players SET hp = ?, attack = ?, defense = ?, agility = ?, gold = ?, stage = ?, xp = ?, level = ?, maxhp = ? WHERE id = ?`
	_, err := a.DB.Exec(query, player.HP, player.Attack, player.Defense, player.Agility,
		player.Gold, player.Stage, player.XP, player.Level, player.MaxHP, player.ID)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) updatePlayerHP(playerID int, newHP int) error {
	query := `UPDATE players SET hp = ? WHERE id = ?`
	_, err := a.DB.Exec(query, newHP, playerID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении HP игрока: %v", err)
	}
	return nil
}

// вычитаем сумму у игрока
func (a *App) updatePlayerGold(playerID int, summ int) error {
	player := a.getPlayer(int64(playerID))
	query := `UPDATE players SET gold = ? WHERE id = ?`
	player.Gold -= summ
	_, err := a.DB.Exec(query, player.Gold, playerID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении HP игрока: %v", err)
	}
	return nil
}

// увеличиваем золото игрока
func (a *App) plusPlayerGold(playerID int, summ int) error {
	player := a.getPlayer(int64(playerID))
	query := `UPDATE players SET gold = ? WHERE id = ?`
	player.Gold += summ
	_, err := a.DB.Exec(query, player.Gold, playerID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении HP игрока: %v", err)
	}
	return nil
}

// удаление
func (a *App) deletePlayer(playerID int) {
	query := `DELETE FROM players WHERE id = ?`
	_, err := a.DB.Exec(query, playerID)
	if err != nil {
		log.Fatal(err)
	}
}

// классы
func (a *App) getClass(class string) Class {
	query := `SELECT c.name, c.base_hp, c.base_atk, c.base_def, c.agility,
					 s.name, s.name_skill, s.help, s.value_base_hp, s.value_base_atk, s.value_base_def, s.value_agility
			  FROM classes c 
			  LEFT JOIN skills s ON s.name = c.name 
   		  WHERE c.name = ?`
	row := a.DB.QueryRow(query, class)

	var c Class
	err := row.Scan(&c.Name, &c.BaseHP, &c.BaseAtk, &c.BaseDef, &c.Agility,
		&c.Button.Name, &c.Button.NameSkill, &c.Button.Help, &c.Button.BaseHP, &c.Button.BaseAtk, &c.Button.BaseDef,
		&c.Button.Agility)
	if err != nil {
		return c
	}
	return c
}

// достаем один предмет
func (a *App) GetItem(value any) Item {
	var i Item
	row := a.DB.QueryRow("SELECT id, name, stat, value, price FROM items WHERE name = ?", value)
	if row.Err() != nil {
		fmt.Println(row.Err())
	}
	row.Scan(&i.ID, &i.Name, &i.Stat, &i.Value, &i.Price)
	return i
}

// достаем все предметы
func (a *App) GetItems() []Item {
	rows, err := a.DB.Query("SELECT id, name, stat, value, price FROM items")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Stat, &item.Value, &item.Price)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, item)
	}

	return items
}

// чекаем время
func (a *App) TimeDuration(name string) time.Time {
	var dur sql.NullTime
	row := a.DB.QueryRow("SELECT duration FROM chanTicker WHERE name = ?", name)
	if row.Err() != nil {
		fmt.Println(row.Err())
	}
	row.Scan(&dur)

	return dur.Time
}

func (a *App) InsertTime(name string, duration time.Duration) {
	_, err := a.DB.Exec("INSERT INTO chanTicker VALUES (?, ?)", name, time.Now().Add(duration))
	if err != nil {
		log.Fatalln("InsertTime - Exec:", err)
	}
}

func (a *App) DeleteTime() {
	_, err := a.DB.Exec("DELETE FROM chanTicker")
	if err != nil {
		log.Fatalln("DeleteTime - Exec:", err)
	}
}

// .Format("02.01.2006 15:04:05")
func (a *App) UpdateTime(name string, duration time.Duration) {
	_, err := a.DB.Exec("UPDATE chanTicker SET duration = ? WHERE name = ?", time.Now().Add(duration), name)
	if err != nil {
		log.Fatalln("InsertTime - Exec:", err)
	}
}

// добавляем предмет
func (a *App) AddItem(chatID int64, item Item) {
	val := randomValue()
	fmt.Println(item.Value + val)
	price := (rand.Intn(val) / 3) - item.Price
	fmt.Println(rand.Intn(val)/3, price, item.Price)
	i := Item{
		// ID:    item.ID,
		Name:  item.Name,
		Stat:  item.Stat,
		Value: item.Value + val,
		Price: item.Price - rand.Intn(val),
	}
	_, err := a.DB.Exec("INSERT INTO player_item (id, name, stat, value, price) VALUES (?, ?, ?, ?, ?)", chatID,
		i.Name, i.Stat, i.Value, i.Price)
	if err != nil {
		log.Fatalln("AddItem - Exec:", err)
	}
}

func (a *App) PlayerItems(chatID int64) []Item {
	var items []Item
	row, err := a.DB.Query("SELECT idItem, name, stat, value, price FROM player_item WHERE id = ?", chatID)
	if err != nil {
		log.Fatalln("PlayerItems - Query:", err)
	}

	for row.Next() {
		var item Item
		row.Scan(&item.ID, &item.Name, &item.Stat, &item.Value, &item.Price)
		items = append(items, item)
	}
	return items
}

// передаем ИД чата и ИД предмета
func (a *App) deleteItem(chatID int64, idItem int) {
	_, err := a.DB.Exec("DELETE FROM player_item WHERE id = ? AND idItem = ?", chatID, idItem)
	if err != nil {
		log.Fatalln("deleteItem - Exec:", err)
	}
}

// обновляем характеристики игрока
func (a *App) RecalcStats(chatID int64, stat string, value int, plusMinus string) {
	query := fmt.Sprintf("UPDATE players SET %s = %s %s ? WHERE id = ?", stat, stat, plusMinus)
	_, err := a.DB.Exec(query, value, chatID)
	if err != nil {
		log.Fatalln("RecalcStats - Exec:", err)
	}
}

// подсчет одетых предметов количество и структуру (itemID, slot)
func (a *App) PlayerItemCount(chatID int64) (int, []struct{ itemID, slot int }) {
	var count int // , itemID, slot int

	var rows []struct {
		itemID, slot int
	}

	a.DB.QueryRow("SELECT COUNT(*) FROM player_item_equip WHERE player_id = ?", chatID).Scan(&count)

	// var rows []int
	row, err := a.DB.Query(`SELECT item_id, slot FROM player_item_equip WHERE player_id = ?`, chatID)
	if err != nil {
		log.Fatalln("PlayerItemCount:", err)
	}
	for row.Next() {
		var rr struct {
			itemID, slot int
		}
		if err := row.Scan(&rr.itemID, &rr.slot); err != nil {
			log.Println("scan error:", err)
			continue
		}
		rows = append(rows, rr)
	}

	return count, rows
}

// одеть предмет
func (a *App) PlayerItemEquip(playerID, itemID, slot int) {
	query := `INSERT INTO player_item_equip(player_id, item_id, slot) VALUES (?, ?, ?)`
	_, err := a.DB.Exec(query, playerID, itemID, slot)
	if err != nil {
		log.Fatalln("PlayerItemEquip - Exec:", err)
	}
}

// удаляем из player_item_equip
func (a *App) DeleteItem(itemID, playerID int) {
	_, err := a.DB.Exec("DELETE FROM player_item_equip WHERE item_id = ? AND player_id = ?", itemID, playerID)
	if err != nil {
		log.Fatalln("DeleteItem - Exec:", err)
	}
}

// Функция для получения случайного значения с разной вероятностью
func randomValue() int {
	rand.NewSource(time.Now().UnixNano())
	val := rand.Intn(100)

	switch {
	case val < 84: // 70
		return RandVal[0]
	case val < 95: // 25
		return RandVal[1]
	default: // 5
		return RandVal[2]
	}
}

// получаем список монстров и боссов
func (a *App) getMonsters() {
	mobs, err := a.DB.Query("select id, name, hp, atk_power, def, xp, gold, is_boss FROM monsters")
	if err != nil {
		log.Fatalln("getMonsters - Query:", err)
	}
	var monster []*Monster
	for mobs.Next() {
		mob := new(Monster)
		// name, hp, atk_power, def, xp, gold, is_boss
		if err := mobs.Scan(&mob.ID, &mob.Name, &mob.HP, &mob.AtkPower, &mob.Def, &mob.XP, &mob.Gold, &mob.IsBoss); err != nil {
			log.Fatalln("getMonsters - mobs.Next:", err)
		}

		monster = append(monster, mob)
	}
	a.Monsters = monster
	// return monster
}
