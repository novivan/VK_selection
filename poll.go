package main

import (
	"errors"
	"log"

	"github.com/tarantool/go-tarantool"
)

type Poll struct {
	ID       string            `json:"id"`
	Creator  string            `json:"creator"`
	Question string            `json:"question"`
	Options  []string          `json:"options"`
	Votes    map[string]int    `json:"votes"`
	Finished bool              `json:"finished"`
}

func NewPoll(id, creator, question string, options []string) *Poll {
	votes := make(map[string]int)
	for _, opt := range options {
		votes[opt] = 0
	}
	return &Poll{
		ID:       id,
		Creator:  creator,
		Question: question,
		Options:  options,
		Votes:    votes,
		Finished: false,
	}
}

// Create сохраняет голосование в Tarantool
func (p *Poll) Create(conn *tarantool.Connection) error {
	// Вставляем голосование в пространство "polls"
	_, err := conn.Insert("polls", []interface{}{p.ID, p.Creator, p.Question, p.Options, p.Votes, p.Finished})
	if err != nil {
		log.Printf("Ошибка при создании голосования в Tarantool: %v", err)
		return err
	}
	return nil
}

// VotePoll регистрирует голос за вариант
func VotePoll(conn *tarantool.Connection, pollID, option string) error {
	// Извлекаем голосование
	resp, err := conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{pollID})
	if err != nil || len(resp.Data) == 0 {
		return errors.New("голосование не найдено")
	}
	// ...разбор ответа и обновление данных...
	// Для упрощения предположим, что мы обновляем поле votes с помощью Update.
	// Необходимо проверить, что вариант существует.
	// Здесь вызывается update-операция Tarantool.
	_, err = conn.Call("vote_func", []interface{}{pollID, option})
	if err != nil {
		log.Printf("Ошибка при голосовании: %v", err)
		return err
	}
	return nil
}

// GetPoll извлекает голосование по ID
func GetPoll(conn *tarantool.Connection, pollID string) (*Poll, error) {
	resp, err := conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{pollID})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("голосование не найдено")
	}
	// Преобразование данных (упрощено)
	data := resp.Data[0].([]interface{})
	poll := &Poll{
		ID:       data[0].(string),
		Creator:  data[1].(string),
		Question: data[2].(string),
		Options:  toStringSlice(data[3]),
		Votes:    toMapStringInt(data[4]),
		Finished: data[5].(bool),
	}
	return poll, nil
}

// FinishPoll завершает голосование, если creator совпадает
func FinishPoll(conn *tarantool.Connection, pollID, creator string) error {
	poll, err := GetPoll(conn, pollID)
	if err != nil {
		return err
	}
	if poll.Creator != creator {
		return errors.New("только создатель может завершить голосование")
	}
	_, err = conn.Call("finish_poll", []interface{}{pollID})
	if err != nil {
		log.Printf("Ошибка при завершении голосования: %v", err)
		return err
	}
	return nil
}

// DeletePoll удаляет голосование из системы
func DeletePoll(conn *tarantool.Connection, pollID string) error {
	_, err := conn.Call("delete_poll", []interface{}{pollID})
	if err != nil {
		log.Printf("Ошибка при удалении голосования: %v", err)
		return err
	}
	return nil
}

// Вспомогательные функции для преобразования типов
func toStringSlice(i interface{}) []string {
	// ...existing code...
	// Пример простой реализации:
	if s, ok := i.([]interface{}); ok {
		var res []string
		for _, v := range s {
			res = append(res, v.(string))
		}
		return res
	}
	return nil
}

func toMapStringInt(i interface{}) map[string]int {
	// ...existing code...
	// Пример простой реализации:
	res := make(map[string]int)
	if m, ok := i.(map[interface{}]interface{}); ok {
		for key, val := range m {
			res[key.(string)] = int(val.(int64))
		}
	} else if m, ok2 := i.(map[string]interface{}); ok2 {
		for key, val := range m {
			res[key] = int(val.(float64))
		}
	}
	return res
}
