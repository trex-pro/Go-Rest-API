package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"project-api/internal/api/helpers"
	"project-api/internal/models"
	"project-api/internal/repositories/sqlconnect"
	"strconv"
)

func GETExecsHandler(w http.ResponseWriter, r *http.Request) {
	var execs []models.Exec
	execs, err := sqlconnect.GETExecsDBHandler(execs, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "SUCCESS",
		Count:  len(execs),
		Data:   execs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GETExecByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handling Path Param.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	exec, err := sqlconnect.GETExecByIDDBHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func POSTExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Exec
	var rawExecs []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Reading Request", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	fields := helpers.GetFieldNames(models.Exec{})

	allowedfields := make(map[string]struct{})
	for _, field := range fields {
		allowedfields[field] = struct{}{}
	}

	for _, exec := range rawExecs {
		for key := range exec {
			_, ok := allowedfields[key]
			if !ok {
				http.Error(w, "Only use for Allowed Fields.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	for _, exec := range newExecs {
		err := helpers.CheckingBlankFields(exec)
		if err != nil {
			return
		}
	}

	addedExecs, err := sqlconnect.POSTExecsDBHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "SUCCESS",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(resp)
}

func PATCHExecsHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]any
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PATCHExecsDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PATCHExecByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Exec ID", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	updatedExec, err := sqlconnect.PATCHExecByIDDBHandler(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedExec)
}

func DELETEExecByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Exec ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DELETEExecByIDDBHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "DELETED",
		ID:     id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
