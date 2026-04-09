package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"project-api/internal/models"
	"project-api/pkg/utils"
	"reflect"
	"strconv"
	"strings"
)

func GETStudentsDBHandler(students []models.Student, r *http.Request) ([]models.Student, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM students WHERE 1=1")
	var args []any

	args = utils.GetTeacherFilter(r, &queryBuilder)
	utils.GetTeacherSort(r, &queryBuilder)

	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer db.Close()

	query := queryBuilder.String()

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class); err != nil {
			return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
		}
		students = append(students, student)
	}
	return students, nil
}

func GETStudentByIDDBHandler(id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRow("SELECT * FROM students WHERE id = ?", id).Scan(
		&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Student{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	} else if err != nil {
		log.Printf("Error: %v", err)
		return models.Student{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	return student, nil
}

func POSTStudentDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		// resp, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newStudent)
		resp, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		lastID, err := resp.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		newStudent.ID = int(lastID)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}

func PUTStudentDBHandler(id int, updatedStudent models.Student) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT * FROM students WHERE id = ?", id).Scan(
		&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedStudent.FirstName, updatedStudent.LastName, updatedStudent.Email, updatedStudent.Class, id)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	return updatedStudent, nil
}

func PATCHStudentsDBHandler(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid ID.")
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid ID.")
		}

		var studentFromDB models.Student
		err = db.QueryRow("SELECT * FROM students WHERE id = ?", id).Scan(
			&studentFromDB.ID, &studentFromDB.FirstName, &studentFromDB.LastName, &studentFromDB.Email, &studentFromDB.Class)
		if err != nil {
			tx.Rollback()
			if err != sql.ErrNoRows {
				return utils.ErrorHandler(err, "Student Not Found in DB.")
			}
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
		}

		// Applying updates using reflect.
		studentVal := reflect.ValueOf(&studentFromDB).Elem()
		studentType := studentVal.Type()

		for key, value := range update {
			if key == "id" {
				continue
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == key+",omitempty" {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(value)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Error: Connot convert %v to %v\n", val.Type(), fieldVal.Type())
							return utils.ErrorHandler(err, "Error Updating Data to DB.")
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ?, subject  = ? WHERE id = ?",
			studentFromDB.FirstName, studentFromDB.LastName, studentFromDB.Email, studentFromDB.Class, studentFromDB.ID)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
		}

		err = tx.Commit()
		if err != nil {
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
		}
	}
	return nil
}

func PATCHStudentByIDDBHandler(id int, updates map[string]any) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT * FROM students WHERE id = ?", id).Scan(
		&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "Student Not Found in DB.")
		}
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()
	for key, value := range updates {
		for i := 0; i < studentType.NumField(); i++ {
			field := studentType.Field(i)
			if field.Tag.Get("json") == key+",omitempty" {
				if studentVal.Field(i).CanSet() {
					fieldVal := studentVal.Field(i)
					fieldVal.Set(reflect.ValueOf(value).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingStudent.FirstName, existingStudent.LastName, existingStudent.Email, existingStudent.Class, id)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	return existingStudent, nil
}

func DELETEStudentsDBHandler(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
		}
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID: %d Not Found.", id))
		}
	}
	return deletedIds, nil
}

func DELETEStudentByIDDBHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Student Not Found in DB.")
	}
	return nil
}
