package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"project-api/internal/api/helpers"
	"project-api/internal/models"
	"project-api/internal/repositories/sqlconnect"
	"project-api/pkg/utils"
	"strconv"
)

func GETStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var students []models.Student
	page, limit := utils.Pagination(r)

	students, totalStudents, err := sqlconnect.GETStudentsDBHandler(students, r, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Status string           `json:"status"`
		Page   int              `json:"page"`
		Limit  int              `json:"limit"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "SUCCESS",
		Page:   page,
		Limit:  limit,
		Count:  totalStudents,
		Data:   students,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GETStudentByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handling Path Param.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	student, err := sqlconnect.GETStudentByIDDBHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func POSTStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var newStudents []models.Student
	var rawStudents []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Reading Request", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &rawStudents)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	fields := helpers.GetFieldNames(models.Student{})

	allowedfields := make(map[string]struct{})
	for _, field := range fields {
		allowedfields[field] = struct{}{}
	}

	for _, student := range rawStudents {
		for key := range student {
			_, ok := allowedfields[key]
			if !ok {
				http.Error(w, "Only use for Allowed Fields.", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newStudents)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	for _, student := range newStudents {
		err := helpers.CheckingBlankFields(student)
		if err != nil {
			return
		}
	}

	addedStudents, err := sqlconnect.POSTStudentDBHandler(newStudents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "SUCCESS",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(resp)
}

func PUTStudentsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}

	var updatedStudent models.Student
	if err := json.NewDecoder(r.Body).Decode(&updatedStudent); err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}
	updatedStudent.ID = id

	updatedStudentFromDB, err := sqlconnect.PUTStudentDBHandler(id, updatedStudent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedStudentFromDB)
}

func PATCHStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]any
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PATCHStudentsDBHandler(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PATCHStudentByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	updatedStudent, err := sqlconnect.PATCHStudentByIDDBHandler(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedStudent)
}

func DELETEStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request Payload.", http.StatusInternalServerError)
		return
	}

	deletedIds, err := sqlconnect.DELETEStudentsDBHandler(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Status     string `json:"status"`
		DeletedIds []int  `json:"deleted_ids"`
	}{
		Status:     "Students deleted successfully",
		DeletedIds: deletedIds,
	}
	json.NewEncoder(w).Encode(resp)
}

func DELETEStudentByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DELETEStudentByIDDBHandler(id)
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
