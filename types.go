package main

import "time"

type CommandCodeRequest struct {
	ID           int
	CommandLogID int
	Command      string
	CommandCode  string
}

type CommandCode struct {
	Command string
	Code    string
}

type CommandLogRequest struct {
	ID        int
	Commands  []string  `json:"commands"` //name to show when serialized to json
	Timestamp time.Time `json:"timestamp"`
}

type CommandLog struct {
	Commands []string `json:"commands"` //name to show when serialized to json
}

type CommandCodeOnly struct {
	CommandCode string `json:"rcr"`
}
