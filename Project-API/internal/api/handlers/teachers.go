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

func GETTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var teachers []models.Teacher
	teachers, err := sqlconnect.GetTeachersDBHandler(teachers, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "SUCCESS",
		Count:  len(teachers),
		Data:   teachers,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GETTeacherByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handling Path Param.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	teacher, err := sqlconnect.GETTeacherByIDDBHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func POSTTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var newTeachers []models.Teacher
	var rawTeachers []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Reading Request", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawTeachers)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	fields := helpers.GetFieldNames(models.Teacher{})

	allowedfields := make(map[string]struct{})
	for _, field := range fields {
		allowedfields[field] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			_, ok := allowedfields[key]
			if !ok {
				http.Error(w, "Only use for Allowed Fields.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newTeachers)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	for _, teacher := range newTeachers {
		err := helpers.CheckingBlankFields(teacher)
		if err != nil {
			return
		}
	}

	addedTeachers, err := sqlconnect.POSTTeacherDBHandler(newTeachers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "SUCCESS",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(resp)
}

func PUTTeachersHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&updatedTeacher); err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}
	updatedTeacher.ID = id

	updatedTeacherFromDB, err := sqlconnect.PUTTeacherDBHandler(id, updatedTeacher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTeacherFromDB)
}

func PATCHTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]any
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PATCHTeachersDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PATCHTeacherByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	updatedTeacher, err := sqlconnect.PATCHTeacherByIDDBHandler(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTeacher)
}

func DELETETeacherHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid  request payload", http.StatusInternalServerError)
		return
	}

	deletedIds, err := sqlconnect.DELETETeacherDBHandler(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Status     string `json:"status"`
		DeletedIds []int  `json:"deleted_ids"`
	}{
		Status:     "Teachers deleted successfully",
		DeletedIds: deletedIds,
	}
	json.NewEncoder(w).Encode(resp)
}

func DELETETeacherByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DELETETeacherByIDDbHandler(id)
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
