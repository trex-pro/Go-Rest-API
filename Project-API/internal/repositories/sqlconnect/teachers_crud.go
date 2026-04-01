package sqlconnect

import (
	"database/sql"
	"log"
	"net/http"
	"project-api/internal/models"
	"reflect"
	"strconv"
	"strings"
)

func GetTeachersDBHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM teachers WHERE 1=1")
	var args []any

	args = getTeacherfilter(r, &queryBuilder)
	getTeacherSort(r, &queryBuilder)

	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer db.Close()

	query := queryBuilder.String()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var teacher models.Teacher
		if err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject); err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
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

func GETTeacherByIDDBHandler(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	} else if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	return teacher, nil
}

func POSTTeacherDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		resp, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}
		lastID, err := resp.LastInsertId()
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func PUTTeacherDBHandler(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, id)
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	return updatedTeacher, nil
}

func PATCHTeachersDBHandler(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return err
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			return err
		}

		var teacherFromDB models.Teacher
		err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDB.ID, &teacherFromDB.FirstName, &teacherFromDB.LastName, &teacherFromDB.Email, &teacherFromDB.Class, &teacherFromDB.Subject)
		if err != nil {
			tx.Rollback()
			if err != sql.ErrNoRows {
				log.Printf("Error: %v", err)
				return err
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
								return err
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
				return err
			}
		}
		err = tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func PATCHTeacherByIDDBHandler(id int, updates map[string]any) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT * FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
	}
	if err != nil {
		log.Printf("Error: %v", err)
		return models.Teacher{}, err
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
		return models.Teacher{}, err
	}
	return existingTeacher, nil
}

func DELETETeacherDBHandler(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		tx.Rollback()
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Printf("Error: %v", err)
			return nil, err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	if len(deletedIds) < 1 {
		return nil, err
	}
	return deletedIds, nil
}

func DELETETeacherByIDDbHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	if rowsAffected == 0 {
		return err
	}
	return nil
}
