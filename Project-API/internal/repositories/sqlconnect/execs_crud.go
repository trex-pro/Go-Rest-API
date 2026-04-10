package sqlconnect

import (
	"database/sql"
	"log"
	"net/http"
	"project-api/internal/models"
	"project-api/pkg/utils"
	"reflect"
	"strconv"
	"strings"
)

func GETExecsDBHandler(Execs []models.Exec, r *http.Request) ([]models.Exec, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM Execs WHERE 1=1")
	var args []any

	args = utils.GetExecFilter(r, &queryBuilder)
	utils.GetExecSort(r, &queryBuilder)

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
		var Exec models.Exec
		err := rows.Scan(&Exec.ID, &Exec.FirstName, &Exec.LastName, &Exec.Email, &Exec.Username, &Exec.UserCreatedAt, &Exec.InactiveStatus, &Exec.Role)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
		}
		Execs = append(Execs, Exec)
	}
	return Execs, nil
}

func GETExecByIDDBHandler(id int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	defer db.Close()

	var Exec models.Exec
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM Execs WHERE id = ?", id).Scan(
		&Exec.ID, &Exec.FirstName, &Exec.LastName, &Exec.Email, &Exec.Username, &Exec.UserCreatedAt, &Exec.InactiveStatus, &Exec.Role)
	if err == sql.ErrNoRows {
		log.Printf("Error: %v", err)
		return models.Exec{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	} else if err != nil {
		log.Printf("Error: %v", err)
		return models.Exec{}, utils.ErrorHandler(err, "Error Retrieving Data from DB.")
	}
	return Exec, nil
}

func POSTExecsDBHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("Execs", models.Exec{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
	}
	defer stmt.Close()

	addedExecs := make([]models.Exec, len(newExecs))
	for i, newExec := range newExecs {
		values := utils.GetStructValues(newExec)
		resp, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		lastID, err := resp.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Adding Data to DB.")
		}
		newExec.ID = int(lastID)
		addedExecs[i] = newExec
	}
	return addedExecs, nil
}

func PATCHExecsDBHandler(updates []map[string]any) error {
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

		var ExecFromDB models.Exec
		err = db.QueryRow("SELECT id, first_name, last_name, email, username, role FROM Execs WHERE id = ?", id).Scan(
			&ExecFromDB.ID, &ExecFromDB.FirstName, &ExecFromDB.LastName, &ExecFromDB.Email, &ExecFromDB.Username, &ExecFromDB.Role)
		if err != nil {
			tx.Rollback()
			if err != sql.ErrNoRows {
				return utils.ErrorHandler(err, "Exec Not Found in DB.")
			}
			return utils.ErrorHandler(err, "Error Updating Data to DB.")
		}

		// Applying updates using reflect.
		ExecVal := reflect.ValueOf(&ExecFromDB).Elem()
		ExecType := ExecVal.Type()

		for key, value := range update {
			if key == "id" {
				continue
			}
			for i := 0; i < ExecVal.NumField(); i++ {
				field := ExecType.Field(i)
				tagName := strings.Split(field.Tag.Get("json"), ",")[0]
				if tagName == key {
					fieldVal := ExecVal.Field(i)
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

		_, err = tx.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ?, role = ? WHERE id = ?",
			ExecFromDB.ID, ExecFromDB.FirstName, ExecFromDB.LastName, ExecFromDB.Email, ExecFromDB.Username, ExecFromDB.Role)
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

func PATCHExecByIDDBHandler(id int, updates map[string]any) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	defer db.Close()

	var existingExec models.Exec
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, role FROM Execs WHERE id = ?", id).Scan(
		&existingExec.ID, &existingExec.FirstName, &existingExec.LastName, &existingExec.Email, &existingExec.Username, &existingExec.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, utils.ErrorHandler(err, "Exec Not Found in DB.")
		}
		return models.Exec{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}

	ExecVal := reflect.ValueOf(&existingExec).Elem()
	ExecType := ExecVal.Type()
	for key, value := range updates {
		for i := 0; i < ExecType.NumField(); i++ {
			field := ExecType.Field(i)
			if field.Tag.Get("json") == key+",omitempty" {
				if ExecVal.Field(i).CanSet() {
					fieldVal := ExecVal.Field(i)
					fieldVal.Set(reflect.ValueOf(value).Convert(ExecVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ?, role = ? WHERE id = ?",
		existingExec.FirstName, existingExec.LastName, existingExec.Email, &existingExec.Username, &existingExec.Role, id)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error Updating Data to DB.")
	}
	return existingExec, nil
}

func DELETEExecByIDDBHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM Execs WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error Deleting Data from DB.")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Exec Not Found in DB.")
	}
	return nil
}
