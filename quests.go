package main

import (
	"fmt"
)

type Quest struct {
	Name     string
	Body     string
	Progress bool
	ID       int
}

func (a *App) getQuest(id int) ([]string, error) {
	var q Quest
	rows, err := a.DB.Query("SELECT * FROM quests WHERE id = ?;", id)
	if err != nil {
		return nil, err
	}

	var arrQ []string
	for rows.Next() {
		rows.Scan(&q.Name, &q.Body, &q.Progress, &q.ID)
		quest := fmt.Sprintf("%s: %s\n", q.Name, q.Body)
		arrQ = append(arrQ, quest)
	}
	return arrQ, nil
}
