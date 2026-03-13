package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project-api/internal/models"
	"project-api/internal/repositories/sqlconnect"
	"reflect"
	"strconv"
	"strings"
)

func TeacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teacherGETHandler(w, r)
	case http.MethodPost:
		teacherPOSTHandler(w, r)
	case http.MethodPut:
		teacherPUTHandler(w, r)
	case http.MethodPatch:
		teacherPATCHHandler(w, r)
	case http.MethodDelete:
		teacherDELETEHandler(w, r)
	default:
		w.Write([]byte("This is TEACHERS route."))
	}
}

func teacherGETHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database:", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimPrefix(path, "/")

	if idStr == "" {
		var queryBuilder strings.Builder
		queryBuilder.WriteString("SELECT * FROM teachers WHERE 1=1")
		var args []any

		args = teacherGETfilter(r, &queryBuilder)
		teacherGETSort(r, &queryBuilder)

		query := queryBuilder.String()

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teachers = make(map[int]models.Teacher)
		teachersList := make([]models.Teacher, 0, len(teachers))
		for rows.Next() {
			var teacher models.Teacher
			if err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject); err != nil {
				log.Printf("Error: %v", err)
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			teachersList = append(teachersList, teacher)
		}

		resp := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "SUCCESS",
			Count:  len(teachersList),
			Data:   teachersList,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Handling Path Param.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		http.Error(w, "Teacher Not Found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func teacherGETfilter(r *http.Request, queryBuilder *strings.Builder) []any {
	var args []any
	params := []string{"first_name", "last_name", "email", "class", "subject"}
	for _, dbField := range params {
		value := r.URL.Query().Get(dbField)
		if value != "" {
			queryBuilder.WriteString(" AND " + dbField + " = ?")
			args = append(args, value)
		}
	}
	return args
}

func teacherGETSort(r *http.Request, queryBuilder *strings.Builder) {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		queryBuilder.WriteString(" ORDER BY")
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) && !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(" " + field + " " + order)
		}
	}
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	vaildFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return vaildFields[field]
}

func teacherPOSTHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database:", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error preparing SQL Query:", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		resp, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Error executing SQL Query:", http.StatusInternalServerError)
			return
		}
		lastID, err := resp.LastInsertId()
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Error retreving Last ID:", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
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

func teacherPUTHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
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

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error retrieving teacher", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, id)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTeacher)
}

func teacherPATCHHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
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

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error retrieving teacher", http.StatusInternalServerError)
		return
	}

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()
	for key, value := range updates {
		for i := 0; i < teacherType.NumField(); i++ {
			field := teacherType.Field(i)
			if field.Tag.Get("json") == key+",omitempty" {
				if teacherVal.Field(i).CanSet() {
					fieldVal := teacherVal.Field(i)
					fieldVal.Set(reflect.ValueOf(value).Convert(teacherVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Email, existingTeacher.Class, existingTeacher.Subject, id)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingTeacher)
}

func teacherDELETEHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Teacher not found", http.StatusNotFound)
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
