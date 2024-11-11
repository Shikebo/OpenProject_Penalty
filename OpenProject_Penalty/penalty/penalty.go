package penalty

import "time"

// Penalty структура для хранения информации о штрафах
type Penalty struct {
	ID        int
	UserName  string
	TaskName  string
	Penalty   float64
	CreatedAt time.Time `json:"created_at"`
}

// CalculatePenalty рассчитывает штраф на основе типа задачи и количества просроченных дней.

// Общая сумма штрафа: базовый штраф + штраф за каждый день просрочки
func CalculatePenalty(taskType string, overdueDays float64) float64 {
	var basePenalty float64

	switch taskType {
	case "Task":
		basePenalty = 10.0 // Базовый штраф за задачу
	case "Bug":
		basePenalty = 20.0 // Базовый штраф за баг
	case "Feature":
		basePenalty = 30.0 // Базовый штраф за фичу
	case "Epic":
		basePenalty = 40.0 // Базовый штраф за эпик
	default:
		basePenalty = 5.0 // Для других типов задач
	}

	totalPenalty := basePenalty * (1 + overdueDays) // Увеличиваем базовый штраф каждый день
	// Общая сумма штрафа: базовый штраф + штраф за каждый день просрочки
	return totalPenalty
}
