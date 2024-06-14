package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
	"io/ioutil"
	"time"
)

type Payload struct {
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

type Response struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Content string `json:"content"`
}

const (
	FileNotFound        = "file not found"
	InternalServerError = "internal server error"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Initializing module")
	if err := createSchemaIfNotExists(ctx, logger, db); err != nil {
		logger.Error("Error creating schema: %v", err)
		return err
	}
	if err := initializer.RegisterRpc("process_file_payload", processFilePayload); err != nil {
		logger.Error("Error registering RPC: %v", err)
		return err
	}
	logger.Info("Module initialized successfully")
	return nil
}

func processFilePayload(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	logger.Info("Processing payload: %s", payload)

	var p Payload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		logger.Error("Error unmarshalling payload: %v", err)
		return "", runtime.NewError(err.Error(), 3)
	}

	// Set default values
	if p.Type == "" {
		p.Type = "core"
	}
	if p.Version == "" {
		p.Version = "1.0.0"
	}
	if p.Hash == "" {
		p.Hash = "null"
	}
	logger.Info("Payload values set: Type=%s, Version=%s, Hash=%s", p.Type, p.Version, p.Hash)

	// Read file
	filePath := fmt.Sprintf("/nakama/data/sample_files/%s/%s.json", p.Type, p.Version)
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Error("Error reading file: %v", err)
		return "", runtime.NewError(FileNotFound, 5)
	}
	logger.Info("File read successfully from path: %s", filePath)

	// Calculate file content hash
	hash := sha256.Sum256(fileContent)
	calculatedHash := hex.EncodeToString(hash[:])
	logger.Info("Calculated hash: %s", calculatedHash)

	// Check if the file already exists in the database
	var existingHash string
	err = db.QueryRowContext(ctx, "SELECT hash FROM file_data WHERE type = $1 AND version = $2", p.Type, p.Version).Scan(&existingHash)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("Error querying database: %v", err)
		return "", runtime.NewError(InternalServerError, 13)
	}
	if existingHash != "" {
		logger.Info("File already exists in the database, skipping save.")
	} else {
		// Save data to the database
		logger.Info("Saving file data to the database")
		_, err = db.Exec("INSERT INTO file_data (type, version, content, hash, processed_at) VALUES ($1, $2, $3, $4, $5)",
			p.Type, p.Version, string(fileContent), calculatedHash, time.Now())
		if err != nil {
			logger.Error("Error saving to database: %v", err)
			return "", runtime.NewError(InternalServerError, 13)
		}
		logger.Info("File data saved to the database successfully")
	}

	// Prepare response
	response := Response{
		Type:    p.Type,
		Version: p.Version,
	}
	if(existingHash == "" || p.Hash == calculatedHash ){
		response.Hash = calculatedHash
		response.Content = string(fileContent)
	} else {
		response.Content = "null"
		response.Hash = "null"
	}
	logger.Info("Response prepared: %v", response)

	res, err := json.Marshal(response)
	if err != nil {
		logger.Error("Error marshalling response: %v", err)
		return "", runtime.NewError(InternalServerError, 13)
	}
	logger.Info("Response marshalled successfully")

	return string(res), nil
}

func createSchemaIfNotExists(ctx context.Context, logger runtime.Logger, db *sql.DB) error {
	logger.Info("Creating schema if not exists")
	_, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS file_data (
            type TEXT,
            version TEXT,
            hash TEXT,
            content TEXT,
            processed_at TIMESTAMP,
            PRIMARY KEY (type, version)
        )
    `)
	if err != nil {
		logger.Error("Error creating schema: %v", err)
	} else {
		logger.Info("Schema created or already exists")
	}
	return err
}
