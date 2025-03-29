package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Обработчик для интеграции с Mattermost
func setupMattermostIntegration() {
	http.HandleFunc("/mattermost/command", handleMattermostCommand)
}

// Обработка slash-команд из Mattermost
func handleMattermostCommand(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondToMattermost(w, "Ошибка обработки команды")
		return
	}

	command := r.Form.Get("command")
	text := r.Form.Get("text")
	userId := r.Form.Get("user_id")

	// Разбираем команду
	switch command {
	case "/poll":
		handlePollCommand(w, text, userId)
	default:
		respondToMattermost(w, "Неизвестная команда")
	}
}

// Обработка команды /poll
func handlePollCommand(w http.ResponseWriter, text, userId string) {
	// Пример разбора команды "/poll create Вопрос? Вариант1, Вариант2, Вариант3"
	// Реальная реализация должна содержать более сложную логику разбора команд
	
	// Отправляем ответ обратно в Mattermost
	respondToMattermost(w, "Команда получена, обрабатывается...")
}

// Отправка ответа в Mattermost
func respondToMattermost(w http.ResponseWriter, text string) {
	response := map[string]string{
		"response_type": "in_channel", // или "ephemeral" для личных сообщений
		"text":          text,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
