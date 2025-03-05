package main

import (
	"fmt"
)

// id INTEGER PRIMARY KEY AUTOINCREMENT,
// name TEXT NOT NULL,
// description TEXT NOT NULL,
// reward TEXT NOT NULL,
// status TEXT NOT NULL,
// required_items TEXT
type Quest struct {
	ID             int
	Name           string
	Description    string
	Reward         string
	Status         string
	Required_items string
}

func (a *App) getQuest(id int) ([]string, error) {
	var q Quest
	rows, err := a.DB.Query("SELECT * FROM quests WHERE id = ?;", id)
	if err != nil {
		return nil, err
	}

	var arrQ []string
	for rows.Next() {
		rows.Scan(&q.Name, &q.ID, &q.Description, &q.Name)
		quest := fmt.Sprintf("%s: %s\n", q.Name, q.Description)
		arrQ = append(arrQ, quest)
	}
	return arrQ, nil
}
