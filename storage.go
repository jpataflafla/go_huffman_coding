package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// temporary solution - no db behavior specified
// and this is for testing/demonstration purposes only
const MaxNumOfLogsInDB = 100

type Storage interface {
	SetCommandLog(*CommandLog) (*CommandLogRequest, error)
	GetAllCommandLogs() ([]*CommandLogRequest, error)
	GetAllCommandCodes() ([]CommandCodeRequest, error)
	GetLatestCommandLog() (*CommandLogRequest, error)
	GetCommandCodesForCommandLog(commandLogID int) ([]CommandCodeRequest, error)
	SetCommandCodes(codes []CommandCode, commandLogID int) ([]CommandCodeRequest, error)
}

type SimplePostgresDB struct {
	db *sql.DB
}

func NewSimplePostgressDB() (*SimplePostgresDB, error) {
	connStr := `user=postgres dbname=postgres sslmode=disable`
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SimplePostgresDB{db: db}, nil
}

func (db *SimplePostgresDB) Init() error {
	if err := db.createCommandLogTable(); err != nil {
		return err
	}

	if err := db.createCommandCodeTable(); err != nil {
		return err
	}

	return nil
}

// temporary solution - no db behavior specified
// and this is for testing/demonstration purposes only
func (db *SimplePostgresDB) DropCommandLogEntriesIfTooMany() error {
	query := "SELECT COUNT(*) FROM CommandLog;"
	var numRows int

	if err := db.db.QueryRow(query).Scan(&numRows); err != nil {
		return err
	}

	if numRows <= MaxNumOfLogsInDB {
		return nil
	}

	if _, err := db.db.Exec("DROP TABLE IF EXISTS CommandLog CASCADE;"); err != nil {
		log.Println("Error dropping CommandLog table:", err)
		return err
	}

	// Recreate the CommandLog table
	if err := db.createCommandLogTable(); err != nil {
		return err
	}

	if _, err := db.db.Exec("DROP TABLE IF EXISTS CommandCode CASCADE;"); err != nil {
		log.Println("Error dropping CommandLog table:", err)
		return err
	}

	// Recreate the CommandCode table
	if err := db.createCommandCodeTable(); err != nil {
		return err
	}

	return nil
}

func (db *SimplePostgresDB) SetCommandLog(commandsLog *CommandLog) (*CommandLogRequest, error) {

	// temp solution for demo purposes
	if err := db.DropCommandLogEntriesIfTooMany(); err != nil {
		return nil, err
	}

	timestamp := time.Now()

	// Marshal the CommandsLogWithTimestamp struct to JSON
	commandsJSON, err := json.Marshal(commandsLog)
	if err != nil {
		return nil, err
	}

	// Execute the SQL query to insert data into the database
	query := "INSERT INTO CommandLog (commands, timestamp) VALUES ($1::JSONB, $2) RETURNING id;"
	row := db.db.QueryRow(query, commandsJSON, timestamp)

	commandsLogWithTimestamp := &CommandLogRequest{
		ID:        -1,
		Commands:  commandsLog.Commands,
		Timestamp: timestamp,
	}
	// Retrieve the ID from the inserted row
	if err := row.Scan(&commandsLogWithTimestamp.ID); err != nil {
		return nil, err
	}

	// Return the created CommandsLogWithTimestamp
	return commandsLogWithTimestamp, nil
}

func (db *SimplePostgresDB) GetAllCommandLogs() ([]*CommandLogRequest, error) {
	query := "SELECT * FROM CommandLog;"

	rows, err := db.db.Query(query)
	if err != nil {
		log.Println("Error querying CommandLog table:", err)
		return nil, err
	}
	defer rows.Close()

	var commandLogsWithTimestamp []*CommandLogRequest

	for rows.Next() {
		var id int
		var commandsJSON []byte
		var timestamp time.Time

		if err := rows.Scan(&id, &commandsJSON, &timestamp); err != nil {
			log.Println("Error scanning row in CommandLog table:", err)
			return nil, err
		}

		var commandLog CommandLog

		// Unmarshal the JSONB field into CommandsLog
		if err := json.Unmarshal(commandsJSON, &commandLog); err != nil {
			log.Println("Error unmarshaling JSON in CommandLog table:", err)
			return nil, err
		}

		commandLogWithTimestamp := CommandLogRequest{
			ID:        id,
			Commands:  commandLog.Commands,
			Timestamp: timestamp,
		}

		commandLogsWithTimestamp =
			append(commandLogsWithTimestamp, &commandLogWithTimestamp)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows in CommandLog table:", err)
		return nil, err
	}

	return commandLogsWithTimestamp, nil
}

func (db *SimplePostgresDB) GetAllCommandCodes() ([]CommandCodeRequest, error) {
	query := "SELECT id, commandLogID, command, commandCode FROM CommandCode;"
	rows, err := db.db.Query(query)
	if err != nil {
		log.Println("Error querying CommandCode table:", err)
		return nil, err
	}
	defer rows.Close()

	var commandCodes []CommandCodeRequest

	for rows.Next() {
		var cc CommandCodeRequest
		if err := rows.Scan(&cc.ID, &cc.CommandLogID, &cc.Command, &cc.CommandCode); err != nil {
			log.Println("Error scanning row from CommandCode table:", err)
			return nil, err
		}

		// Convert CommandCodeRequest to CommandCode
		commandCodes = append(commandCodes, cc)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows from CommandCode table:", err)
		return nil, err
	}

	return commandCodes, nil
}

func (db *SimplePostgresDB) GetLatestCommandLog() (*CommandLogRequest, error) {
	// Get the latest CommandLog id
	latestCommandLogQuery := "SELECT id, commands, timestamp FROM CommandLog ORDER BY timestamp DESC LIMIT 1;"
	commandLogRow := db.db.QueryRow(latestCommandLogQuery)

	var latestCommandLog CommandLogRequest
	var commandsJSON []byte

	if err := commandLogRow.Scan(&latestCommandLog.ID, &commandsJSON, &latestCommandLog.Timestamp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Handle case where no rows were found
			log.Println("No matching CommandLog found")
			return nil, err
		}

		log.Println("Error scanning row from CommandLog table:", err)
		return nil, err
	}

	// Convert the JSON array to a string slice
	var commandLog CommandLog
	// Unmarshal the JSONB field into CommandsLog
	if err := json.Unmarshal(commandsJSON, &commandLog); err != nil {
		log.Println("Error unmarshaling JSON in CommandLog table:", err)
		return nil, err
	}
	latestCommandLog.Commands = commandLog.Commands

	return &latestCommandLog, nil
}

func (db *SimplePostgresDB) GetCommandCodesForCommandLog(commandLogID int) ([]CommandCodeRequest, error) {
	// Now, get CommandCode rows for the latest CommandLog
	commandCodeQuery := "SELECT id, commandLogID, command, commandCode FROM CommandCode WHERE commandLogID = $1;"
	rows, err := db.db.Query(commandCodeQuery, commandLogID)
	if err != nil {
		log.Println("Error querying CommandCode table:", err)
		return nil, err
	}
	defer rows.Close()

	var commandCodes []CommandCodeRequest

	for rows.Next() {
		var cc CommandCodeRequest
		if err := rows.Scan(&cc.ID, &cc.CommandLogID, &cc.Command, &cc.CommandCode); err != nil {
			log.Println("Error scanning row from CommandCode table:", err)
			return nil, err
		}
		commandCodes = append(commandCodes, cc)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows from CommandCode table:", err)
		return nil, err
	}

	// Return commandCodes, whether it's empty or not
	return commandCodes, nil
}

func (db *SimplePostgresDB) SetCommandCodes(codes []CommandCode, commandLogID int) ([]CommandCodeRequest, error) {
	// Create a slice to store the inserted command codes
	var insertedCodes []CommandCodeRequest
	log.Println(codes)
	// Iterate through the provided codes and insert each one into the CommandCode table
	for _, code := range codes {
		// Execute the SQL query to insert data into the CommandCode table
		query := "INSERT INTO CommandCode (commandLogID, command, commandCode) VALUES ($1, $2, $3) RETURNING id, commandLogID, command, commandCode;"
		row := db.db.QueryRow(query, commandLogID, code.Command, code.Code)

		// Create a CommandCode instance to store the inserted code details
		var insertedCode CommandCodeRequest
		// Retrieve the ID from the inserted row
		if err := row.Scan(&insertedCode.ID, &insertedCode.CommandLogID, &insertedCode.Command, &insertedCode.CommandCode); err != nil {
			return nil, err
		}

		// Append the inserted code to the slice
		insertedCodes = append(insertedCodes, insertedCode)
	}

	// Return the slice of inserted command codes
	return insertedCodes, nil
}

//Create tables

func (db *SimplePostgresDB) createCommandLogTable() error {
	//JSONB uses more memory, but may be more future-proof than TEXT[] - in case the input format changes
	query := `
		CREATE TABLE IF NOT EXISTS CommandLog (
			id serial PRIMARY KEY,
			commands JSONB NOT NULL,
			timestamp TIMESTAMP
		);
	`

	if _, err := db.db.Exec(query); err != nil {
		log.Println("Error creating CommandLog table:", err)
		return err
	}

	return nil
}

func (db *SimplePostgresDB) createCommandCodeTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS CommandCode (
			id serial PRIMARY KEY,
			commandLogID INT REFERENCES CommandLog(id) ON DELETE CASCADE,
			command TEXT,
			commandCode TEXT
		);
	`

	if _, err := db.db.Exec(query); err != nil {
		log.Println("Error creating CommandCode table:", err)
		return err
	}

	return nil
}
