package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/grantfbarnes/card-judge/auth"
)

type User struct {
	Id            uuid.UUID
	CreatedOnDate time.Time
	ChangedOnDate time.Time

	Name         string
	PasswordHash string
	ColorTheme   sql.NullString
	IsApproved   bool
	IsAdmin      bool
}

func SearchUsers(search string) ([]User, error) {
	if search == "" {
		search = "%"
	}

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_APPROVED,
			IS_ADMIN
		FROM USER
		WHERE NAME LIKE ?
		ORDER BY
			CHANGED_ON_DATE DESC,
			NAME ASC
	`
	rows, err := query(sqlString, search)
	if err != nil {
		return nil, err
	}

	result := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.Id,
			&user.CreatedOnDate,
			&user.ChangedOnDate,
			&user.Name,
			&user.PasswordHash,
			&user.ColorTheme,
			&user.IsApproved,
			&user.IsAdmin); err != nil {
			log.Println(err)
			return result, errors.New("failed to scan row in query results")
		}
		result = append(result, user)
	}
	return result, nil
}

func GetUser(userId uuid.UUID) (User, error) {
	var user User

	sqlString := `
		SELECT
			ID,
			CREATED_ON_DATE,
			CHANGED_ON_DATE,
			NAME,
			PASSWORD_HASH,
			COLOR_THEME,
			IS_APPROVED,
			IS_ADMIN
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return user, err
	}

	for rows.Next() {
		if err := rows.Scan(
			&user.Id,
			&user.CreatedOnDate,
			&user.ChangedOnDate,
			&user.Name,
			&user.PasswordHash,
			&user.ColorTheme,
			&user.IsApproved,
			&user.IsAdmin); err != nil {
			log.Println(err)
			return user, errors.New("failed to scan row in query results")
		}
	}

	return user, nil
}

func GetUserPasswordHash(userId uuid.UUID) (string, error) {
	var passwordHash string

	sqlString := `
		SELECT
			PASSWORD_HASH
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return passwordHash, err
	}

	for rows.Next() {
		if err := rows.Scan(&passwordHash); err != nil {
			log.Println(err)
			return passwordHash, errors.New("failed to scan row in query results")
		}
	}

	return passwordHash, nil
}

func GetUserIsApproved(userId uuid.UUID) (bool, error) {
	var isApproved bool

	sqlString := `
		SELECT
			IS_APPROVED
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return isApproved, err
	}

	for rows.Next() {
		if err := rows.Scan(&isApproved); err != nil {
			log.Println(err)
			return isApproved, errors.New("failed to scan row in query results")
		}
	}

	return isApproved, nil
}

func GetUserIsAdmin(userId uuid.UUID) (bool, error) {
	var isAdmin bool

	sqlString := `
		SELECT
			IS_ADMIN
		FROM USER
		WHERE ID = ?
	`
	rows, err := query(sqlString, userId)
	if err != nil {
		return isAdmin, err
	}

	for rows.Next() {
		if err := rows.Scan(&isAdmin); err != nil {
			log.Println(err)
			return isAdmin, errors.New("failed to scan row in query results")
		}
	}

	return isAdmin, nil
}

func AddUserLoginAttempt(ipAddress string, userName string) error {
	sqlString := `
		INSERT INTO LOGIN_ATTEMPT (IP_ADDRESS, USER_NAME)
		VALUES (?, ?)
	`
	return execute(sqlString, ipAddress, userName)
}

func AllowUserLoginAttempt(ipAddress string, userName string) (bool, error) {
	sqlString := "SELECT FN_GET_LOGIN_ATTEMPT_IS_ALLOWED (?, ?)"
	rows, err := query(sqlString, ipAddress, userName)
	if err != nil {
		return false, err
	}

	allowLogin := false
	for rows.Next() {
		if err := rows.Scan(&allowLogin); err != nil {
			log.Println(err)
			return false, errors.New("failed to scan row in query results")
		}
	}

	return allowLogin, nil
}

func GetUserIdByName(name string) (uuid.UUID, error) {
	var userId uuid.UUID

	sqlString := `
		SELECT
			ID
		FROM USER
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return userId, err
	}

	for rows.Next() {
		if err := rows.Scan(&userId); err != nil {
			log.Println(err)
			return userId, errors.New("failed to scan row in query results")
		}
	}

	if userId == uuid.Nil {
		return userId, errors.New("user not found")
	}

	return userId, nil
}

func UserNameExists(name string) bool {
	sqlString := `
		SELECT
			ID
		FROM USER
		WHERE NAME = ?
	`
	rows, err := query(sqlString, name)
	if err != nil {
		return false
	}

	return rows.Next()
}

func CreateUser(name string, password string, isApproved bool) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		INSERT INTO USER (NAME, PASSWORD_HASH, IS_APPROVED)
		VALUES (?, ?, ?)
	`
	return execute(sqlString, name, passwordHash, isApproved)
}

func ApproveUser(id uuid.UUID) error {
	sqlString := `
		UPDATE USER
		SET IS_APPROVED = 1
		WHERE ID = ?
	`
	return execute(sqlString, id)
}

func SetUserName(id uuid.UUID, name string) error {
	sqlString := `
		UPDATE USER
		SET
			NAME = ?
		WHERE ID = ?
	`
	err := execute(sqlString, name, id)
	if err != nil {
		return err
	}

	return nil
}

func SetUserPassword(id uuid.UUID, password string) error {
	passwordHash, err := auth.GetPasswordHash(password)
	if err != nil {
		log.Println(err)
		return errors.New("failed to hash password")
	}

	sqlString := `
		UPDATE USER
		SET
			PASSWORD_HASH = ?
		WHERE ID = ?
	`
	err = execute(sqlString, passwordHash, id)
	if err != nil {
		return err
	}

	return nil
}

func SetUserColorTheme(id uuid.UUID, colorTheme string) error {
	sqlString := `
		UPDATE USER
		SET
			COLOR_THEME = ?
		WHERE ID = ?
	`
	var err error
	if colorTheme == "" {
		err = execute(sqlString, nil, id)
	} else {
		err = execute(sqlString, colorTheme, id)
	}
	if err != nil {
		return err
	}

	return nil
}

func SetUserIsAdmin(id uuid.UUID, isAdmin bool) error {
	sqlString := `
		UPDATE USER
		SET
			IS_ADMIN = ?
		WHERE ID = ?
	`
	err := execute(sqlString, isAdmin, id)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(id uuid.UUID) error {
	sqlString := `
		DELETE FROM USER
		WHERE ID = ?
	`
	err := execute(sqlString, id)
	if err != nil {
		return err
	}

	return nil
}
