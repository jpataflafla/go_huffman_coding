package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"command-encoding-service/pkg/generate_codes"
)

type simpleAPIServer struct {
	listenAddress string
	storage       Storage
}

type APIError struct {
	Error string
}

func NewApiServer(listenAddress string, storage Storage) *simpleAPIServer {
	return &simpleAPIServer{
		listenAddress: listenAddress,
		storage:       storage,
	}
}

func (s *simpleAPIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/commands", makeHTTPHandlerFunc(s.handleCommands))
	router.HandleFunc("/rcr/{command}", makeHTTPHandlerFunc(s.handleGetCodeForCommandFromLastCommandLog))
	router.HandleFunc("/allCommandCodes", makeHTTPHandlerFunc(s.handleGetAllCommandCodes))

	log.Println("JSON API server running on port: ", s.listenAddress)
	http.ListenAndServe(s.listenAddress, router)
}

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

type apiHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func makeHTTPHandlerFunc(apiHandlerFunc apiHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := apiHandlerFunc(w, r); err != nil {
			// error handling
			writeJson(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func (s *simpleAPIServer) handleCommands(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAllCommandLogs(w)
	}
	if r.Method == "POST" {
		return s.handlePostCommands(w, r)
	}

	return fmt.Errorf("request method not allowed: %s", r.Method)
}

func (s *simpleAPIServer) handleGetCodeForCommandFromLastCommandLog(w http.ResponseWriter, r *http.Request) error {
	command := mux.Vars(r)["command"]

	// get code from DB or memory and send code
	code, err := getCodeForCommandFromLastCommandLog(command, s.storage)
	if err != nil {
		// Check if the error is due to the command not being found
		if errors.Is(err, ErrCommandNotFound) {
			// Return a 404 Not Found response
			http.Error(w, "Command not found", http.StatusNotFound)
			return nil
		}

		// Handle other errors
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	// Construct the response
	commandCode := CommandCodeOnly{CommandCode: code}
	return writeJson(w, http.StatusOK, commandCode)
}

func (s *simpleAPIServer) handlePostCommands(w http.ResponseWriter, r *http.Request) error {
	commandsLog := &CommandLog{}
	if err := json.NewDecoder(r.Body).Decode(commandsLog); err != nil {
		return err
	}

	commandsLogWithTimestamp, err := s.storage.SetCommandLog(commandsLog)
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, commandsLogWithTimestamp)
}

func (s *simpleAPIServer) handleGetAllCommandLogs(w http.ResponseWriter) error { //, r *http.Request) error {
	// Call the storage method to get all command logs
	commandLogs, err := s.storage.GetAllCommandLogs()
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, commandLogs)
}

func (s *simpleAPIServer) handleGetAllCommandCodes(w http.ResponseWriter, r *http.Request) error {
	allCommandCodes, err := s.storage.GetAllCommandCodes()
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, allCommandCodes)
}

// Define a custom error type for command not found
var ErrCommandNotFound = errors.New("command not found")

func getCodeForCommandFromLastCommandLog(command string, db Storage) (string, error) {
	commandLog, err := db.GetLatestCommandLog()
	if err != nil {
		return "", err
	}
	//commands := []string{"LEFT", "GRAB", "LEFT", "BACK", "LEFT", "BACK", "LEFT"}
	comandCodes, err := db.GetCommandCodesForCommandLog(commandLog.ID)
	if err != nil {
		return "", err
	}

	if len(comandCodes) == 0 {
		// generate codes using command log
		codeMap := generate_codes.GetCodesFromListOfCommands(commandLog.Commands)
		codes := ConvertCodesToCommandCodeSlice(codeMap)
		comandCodes, err = db.SetCommandCodes(codes, commandLog.ID)
		if err != nil {
			return "", err
		}
	}

	for _, code := range comandCodes {
		if code.Command == command {
			return code.CommandCode, nil
		}
	}

	// If the command is not found, return a custom error
	return "", ErrCommandNotFound
}

func ConvertCodesToCommandCodeSlice(codes map[string]string) []CommandCode {
	commandCodes := make([]CommandCode, 0, len(codes))
	for cmd, code := range codes {
		commandCodes = append(commandCodes, CommandCode{Command: cmd, Code: code})
	}
	return commandCodes
}
