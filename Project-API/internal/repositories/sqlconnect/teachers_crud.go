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

func GETTeachersDBHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM teachers WHERE 1=1")
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
		var teacher models.Teacher
		if err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject); err != nil {
			return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

func GETTeacherByIDDBHandler(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Teacher{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	} else if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	return teacher, nil
}

func POSTTeacherDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		// resp, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newTeacher)
		resp, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		lastID, err := resp.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func PUTTeacherDBHandler(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, id)
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	return updatedTeacher, nil
}

func PATCHTeachersDBHandler(updates []map[string]any) error {
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

		var teacherFromDB models.Teacher
		err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDB.ID, &teacherFromDB.FirstName, &teacherFromDB.LastName, &teacherFromDB.Email, &teacherFromDB.Class, &teacherFromDB.Subject)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Teacher Not Found in DB.")
			}
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
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
							return utils.ErrorHandler(fmt.Errorf("cannot convert %v to %v", val.Type(), fieldVal.Type()), "Error Updating Data to DB.")
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
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
		}
	}
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	return nil
}

func PATCHTeacherByIDDBHandler(id int, updates map[string]any) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.ErrorHandler(err, "Teacher Not Found in DB.")
		}
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
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
		return models.Teacher{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	return existingTeacher, nil
}

func DELETETeachersDBHandler(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
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

func DELETETeacherByIDDBHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Teacher Not Found in DB.")
	}
	return nil
}

func GETStudentsByTeacherIDDBHandler(teacherID string) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer db.Close()

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT id, first_name, last_name, email, class FROM students WHERE class = (SELECT class from teachers WHERE id = ?)")
	query := queryBuilder.String()
	rows, err := db.Query(query, teacherID)
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			log.Println(err)
			return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
		}
		students = append(students, student)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	return students, nil
}
