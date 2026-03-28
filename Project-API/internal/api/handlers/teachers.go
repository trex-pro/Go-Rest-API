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

func GETTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database:", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM teachers WHERE 1=1")
	var args []any

	args = getTeacherfilter(r, &queryBuilder)
	getTeacherSort(r, &queryBuilder)

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
}

func GETTeacherByIDHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database:", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	idStr := r.PathValue("id")

	// Handling Path Param.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
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

func getTeacherfilter(r *http.Request, queryBuilder *strings.Builder) []any {
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

func getTeacherSort(r *http.Request, queryBuilder *strings.Builder) {
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

func POSTTeachersHandler(w http.ResponseWriter, r *http.Request) {
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

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
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

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, id)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTeacher)
}

func PATCHTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var updates []map[string]any
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error in Request Payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error converting String to Int", http.StatusInternalServerError)
			return
		}

		var teacherFromDB models.Teacher
		err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDB.ID, &teacherFromDB.FirstName, &teacherFromDB.LastName, &teacherFromDB.Email, &teacherFromDB.Class, &teacherFromDB.Subject)
		if err != nil {
			tx.Rollback()
			if err != sql.ErrNoRows {
				log.Printf("Error: %v", err)
				http.Error(w, "Teacher not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error retreving teacher", http.StatusInternalServerError)
			return
		}

		// Applying updates using reflect.
		teacherVal := reflect.ValueOf(&teacherFromDB).Elem()
		teacherType := teacherVal.Type()

		for key, value := range update {
			if key == "id" {
				continue
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == key+",omitempty" {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(value)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Error: Connot convert %v to %v\n", val.Type(), fieldVal.Type())
							return
						}
					}
					break
				}
			}
		}
		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject  = ? WHERE id = ?",
			teacherFromDB.FirstName, teacherFromDB.LastName, teacherFromDB.Email, teacherFromDB.Class, teacherFromDB.Subject, teacherFromDB.ID)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			http.Error(w, "Error updating teacher", http.StatusInternalServerError)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error commitng changes", http.StatusInternalServerError)
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

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Email, existingTeacher.Class, existingTeacher.Subject, id)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingTeacher)
}

func DELETETeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid  request payload", http.StatusInternalServerError)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		tx.Rollback()
		log.Printf("Error: %v", err)
		http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Printf("Error: %v", err)
			http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Error retrieving deleting teacher", http.StatusInternalServerError)
			return
		}
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("%d does not exist", id), http.StatusBadRequest)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error commitng changes", http.StatusInternalServerError)
		return
	}
	if len(deletedIds) < 1 {
		http.Error(w, "IDs do not exist", http.StatusBadRequest)
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
