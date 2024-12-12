package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type App struct {
	DB *sql.DB
}

func (a *App) ConnDB() {
	db, err := sql.Open("sqlite3", "./files.db")
	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
}

func CraeteDataBase() {
	file, err := os.ReadDir("./")
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range file {
		if f.Name() == "files.db" {
			return
		}
	}
	// Подготовка команды sqlite3 files.db
	cmd := exec.Command("sqlite3", "files.db")

	// Запуск команды
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) createTablesAndDB() {
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
		level INTEGER,
		maxhp INTEGER
	);`, `CREATE TABLE IF NOT EXISTS items (
    name TEXT,
    stat TEXT,
    value INTEGER,
    price INTEGER,
    PRIMARY KEY (name)
);
`, `CREATE TABLE IF NOT EXISTS quests (
		name     TEXT,
		body     TEXT,
		progress BOOLEAN,
		ID       INTEGER
		);`, `INSERT INTO "items" (name, stat, value, price) VALUES
("Кольцо Древнего Леса", "HP", 50, 130),
("Меч Огненной Стужи", "Attack", 75, 70),
("Щит Светлого Воина", "Defense", 60, 110),
("Перчатки Ловкости", "Attack", 25, 40),
("Плащ Призрачной Тени", "Defense", 45, 30),
("Серебряный Амулет", "HP", 30, 70),
("Книга Заклинаний", "Attack", 40, 60),
("Сапоги Ускользания", "Defense", 20, 30),
("Талисман Мудрости", "HP", 15, 60),
("Мантия Ледяного Ветра", "Defense", 70, 50),
("Кристалл Души", "HP", 100, 176),
("Кинжал Ядовитой Змеи", "Attack", 50, 77),
("Шлем Легендарного Героя", "Defense", 80, 96),
("Пояс Силы", "Attack", 60, 99),
("Эльфийский Лук", "Attack", 90, 158),
("Сумка Бесконечности", "HP", 20, 60),
("Чаша Мудрости", "Defense", 40, 50),
("Медальон Силы", "HP", 10, 39),
("Перья Летучей Мыши", "Attack", 15, 40);`,
		`CREATE TABLE IF NOT EXISTS chanTicker (
    name TEXT,
    duration DATETIME DEFAULT CURRENT_TIMESTAMP
);`}
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
("Воин", 150, 20, 10, 4),
("Маг", 100, 30, 5, 3),
("Лучник", 120, 25, 7, 4),
("Некромант", 110, 28, 6, 3),
("Жрец", 90, 15, 8, 3),
("Паладин", 140, 22, 12, 3),
("Тёмный Рыцарь", 130, 27, 9, 3);`,
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
	query := `INSERT INTO players (id, name, class, hp, attack, defense, gold, stage, xp, level, maxhp) VALUES (?, ?, ?, ?, ?, ?, 0, 1, 0, 1, ?)`
	_, err := a.DB.Exec(query, userID, name, class, playerClass.BaseHP, playerClass.BaseAtk, playerClass.BaseDef, playerClass.BaseHP)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) getPlayer(userID int64) *Player {
	query := `SELECT id, name, class, hp, attack, defense, gold, stage, xp, level, maxhp FROM players WHERE id = ?`
	row := a.DB.QueryRow(query, userID)

	var player Player
	err := row.Scan(&player.ID, &player.Name, &player.Class, &player.HP, &player.Attack, &player.Defense,
		&player.Gold, &player.Stage, &player.XP, &player.Level, &player.MaxHP)
	if err != nil {
		return nil
	}
	return &player
}

func (a *App) updatePlayer(player *Player) {
	query := `UPDATE players SET hp = ?, attack = ?, defense = ?, gold = ?, stage = ?, xp = ?, level = ?, maxhp = ? WHERE id = ?`
	_, err := a.DB.Exec(query, player.HP, player.Attack, player.Defense,
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
func (a *App) GetItem(name string) Item {
	var i Item
	row := a.DB.QueryRow("SELECT name, stat, value, price FROM items WHERE name = ?", name)
	if row.Err() != nil {
		fmt.Println(row.Err())
	}
	row.Scan(&i.Name, &i.Stat, &i.Value, &i.Price)
	return i
}

// достаем все предметы
func (a *App) GetItems() []Item {
	var i Item
	rows, err := a.DB.Query("SELECT name, stat, value, price FROM items")
	if err != nil {
		log.Fatalln("GetItems - Query:", err)
	}

	arrItem := []Item{}
	for rows.Next() {
		rows.Scan(&i.Name, &i.Stat, &i.Value, &i.Price)
		arrItem = append(arrItem, i)
	}
	return arrItem
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

// .Format("02.01.2006 15:04:05")
func (a *App) UpdateTime(name string, duration time.Duration) {
	_, err := a.DB.Exec("UPDATE chanTicker SET duration = ? WHERE name = ?", time.Now().Add(duration), name)
	if err != nil {
		log.Fatalln("InsertTime - Exec:", err)
	}
}
