package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool"
)

// Глобальное подключение к Tarantool и конфигурация
var db *tarantool.Connection

func main() {
	var err error
	// Подключение к Tarantool (адрес и порт можно настроить через переменные окружения)
	db, err = tarantool.Connect(os.Getenv("TARANTOOL_URI"), tarantool.Opts{
		User: os.Getenv("TARANTOOL_USER"),
		Pass: os.Getenv("TARANTOOL_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к Tarantool: %v", err)
	}
	log.Println("Подключение к Tarantool установлено")

	// Настройка HTTP-эндпоинтов
	http.HandleFunc("/create", handleCreatePoll)
	http.HandleFunc("/vote", handleVotePoll)
	http.HandleFunc("/results", handlePollResults)
	http.HandleFunc("/finish", handleFinishPoll)
	http.HandleFunc("/delete", handleDeletePoll)

	// Добавляем интеграцию с Mattermost
	setupMattermostIntegration()

	log.Println("Сервер запущен на порту 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}

func handleCreatePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Creator  string   `json:"creator"`
		Question string   `json:"question"`
		Options  []string `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	poll := NewPoll(uuid.New().String(), input.Creator, input.Question, input.Options)
	if err := poll.Create(db); err != nil {
		http.Error(w, "Ошибка создания голосования", http.StatusInternalServerError)
		return
	}
	log.Printf("Голосование создано: %s", poll.ID)
	json.NewEncoder(w).Encode(poll)
}

func handleVotePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		PollID string `json:"poll_id"`
		Option string `json:"option"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	if err := VotePoll(db, input.PollID, input.Option); err != nil {
		http.Error(w, "Ошибка голосования: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Голос получен для %s за вариант %s", input.PollID, input.Option)
	w.Write([]byte("Голос засчитан"))
}

func handlePollResults(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Query().Get("poll_id")
	if pollID == "" {
		http.Error(w, "Отсутствует poll_id", http.StatusBadRequest)
		return
	}
	poll, err := GetPoll(db, pollID)
	if err != nil {
		http.Error(w, "Ошибка получения результатов: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(poll)
}

func handleFinishPoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		PollID  string `json:"poll_id"`
		Creator string `json:"creator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}
	if err := FinishPoll(db, input.PollID, input.Creator); err != nil {
		http.Error(w, "Ошибка завершения голосования: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Голосование %s завершено", input.PollID)
	w.Write([]byte("Голосование завершено"))
}

func handleDeletePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	pollID := r.URL.Query().Get("poll_id")
	if pollID == "" {
		http.Error(w, "Отсутствует poll_id", http.StatusBadRequest)
		return
	}
	if err := DeletePoll(db, pollID); err != nil {
		http.Error(w, "Ошибка удаления голосования: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Голосование %s удалено", pollID)
	w.Write([]byte("Голосование удалено"))
}
