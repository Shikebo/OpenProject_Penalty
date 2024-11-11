package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OverdueTasks struct {
	ID          int     `json:"ID"`
	Subject     string  `json:"Subject"`
	Type        string  `json:"Type"`
	Responsible string  `json:"Responsible"`
	Penalty     float64 `json:"Penalty"`
}

func GetOverdueTasks() ([]OverdueTasks, error) {
	resp, err := http.Get("http://localhost:3000/penalty")
	if err != nil {
		return nil, fmt.Errorf("error fetching notifications: %w", err)
	}

	defer resp.Body.Close()

	var response []OverdueTasks
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response, nil
}

func FormatOverdueTasksForTelegram(tasks []OverdueTasks) string {
	var formattedTasks string
	for _, t := range tasks {
		formattedTasks += fmt.Sprintf(
			"───────────────────────────\n"+
				"🔔 **Уведомление ID:** %d\n"+
				"📂 **Название:** %s\n"+
				"👤 **Ответственный:** %s\n"+
				"📌 **Тип задачи:** %s\n"+
				"💵 **Штраф:** %.2f\n"+
				"───────────────────────────\n",
			t.ID,
			t.Subject,
			t.Responsible,
			t.Type,
			t.Penalty,
		)
	}
	return formattedTasks
}
