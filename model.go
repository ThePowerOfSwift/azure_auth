package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"

	_ "github.com/mattn/go-sqlite3"
)

type UserDataToSave struct {
	UserId            string
	AccessToken       string
	RefreshToken      string
	ClientPublicToken string
	TemporaryToken    string
}

type TokenInterface interface {
	getAccessToken() string
	getRefreshToken() string
	getTemporaryToken() string
	publicToken() string
}

func NewUserDataToSave(o TokenInterface, l LoggedUserInfo) UserDataToSave {
	return UserDataToSave{
		UserId:            l.ID,
		AccessToken:       o.getAccessToken(),
		RefreshToken:      o.getRefreshToken(),
		ClientPublicToken: o.publicToken(),
		TemporaryToken:    o.getTemporaryToken(),
	}
}

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

func CreateTable(db *sql.DB) {
	sqlTable := `
	CREATE TABLE IF NOT EXISTS user_auth_data(
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		UserId TEXT,
		AccessToken TEXT,
		RefreshToken TEXT,
		ClientPublicToken TEXT,
		TemporaryToken TEXT
	);
	`

	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
}

func StoreItem(db *sql.DB, item UserDataToSave) {
	sqlStatement := `
	INSERT OR REPLACE INTO user_auth_data(
		UserId,
		AccessToken,
		RefreshToken,
		ClientPublicToken,
	    TemporaryToken
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(item.UserId, item.AccessToken, item.RefreshToken, item.ClientPublicToken, item.TemporaryToken)
	if err2 != nil {
		panic(err2)
	}
}
func FindRecord(db *sql.DB, token string) (item []UserDataToSave, err error) {
	sqlStatement := `
	SELECT ClientPublicToken, AccessToken, RefreshToken FROM user_auth_data
	WHERE ClientPublicToken = $1
	`

	rows, err := db.Query(sqlStatement, token)
	if err != nil {
		panic(fmt.Errorf("ERROR: %s", err))
	}
	defer rows.Close()

	var result []UserDataToSave
	var count uint32
	for rows.Next() {
		item := UserDataToSave{}
		err = rows.Scan(&item.ClientPublicToken, &item.AccessToken, &item.RefreshToken)
		count++
		result = append(result, item)
	}
	if count == 0 {
		fmt.Println("No records found with public token:", token)
		err = fmt.Errorf("No records found with public token: %s", token)
	}
	return result, err
}

func UpdateItem(db *sql.DB, token, authToken, refreshToken string) (err error) {
	sqlStatement := `
		UPDATE user_auth_data
		SET AccessToken = $1, RefreshToken = $2
		WHERE ClientPublicToken = $3;
	`
	_, err = db.Exec(sqlStatement, authToken, refreshToken, token)

	return err
}

func ChangeTemporaryTokenToPublic(db *sql.DB, token string) (clientPublicToken string, err error) {
	clientPublicToken = fmt.Sprint(uuid.New())
	sqlStatement := `
			UPDATE user_auth_data
			SET ClientPublicToken = $1, TemporaryToken = $2
			WHERE TemporaryToken = $3;
		`

	res, err := db.Exec(sqlStatement, clientPublicToken, "", token)
	if err != nil {
		return "", err
	}
	var rows int64
	rows, _ = res.RowsAffected()
	if rows == 0 {
		fmt.Println("Record with TemporaryToken", token, "not found")
		err = fmt.Errorf("Record with TemporaryToken %s, not found", token)
	}
	return clientPublicToken, err
}
