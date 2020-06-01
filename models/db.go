package models

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

type User struct {
	Username string
	Password string
	Email    string
}

type Vault struct {
	Id       int
	Account  string
	Password string
}

func ConnectDB() *sql.DB {

	// CONNECTION TO DATABASE
	url, _ := os.LookupEnv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	Try(err)

	// PING for FORCING DATABASE TO OPEN A CONNECTION
	err = db.Ping()
	Try(err)

	//Create Table
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS users (
				username varchar(25) NOT NULL PRIMARY KEY,
				password char(64) NOT NULL,
				email varchar(45) NOT NULL
			);
		`)
	Try(err)

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS vault (
				id serial NOT NULL PRIMARY KEY,
				password varchar(60) NOT NULL,
				account varchar(60) NOT NULL,
				username varchar(25) references users(username)
			);
		`)
	Try(err)

	return db
}

// INSERIRE un NUOVO utente
func NewUser(db *sql.DB, u User) {
	_, err := db.Exec(
		`INSERT INTO users(username,password,email) VALUES($1,$2,$3);`,
		u.Username, u.Password, u.Email)
	Try(err)
}

func NewVault(db *sql.DB, username string, key string, account string, password string) {
	password, _ = Encrypt([]byte(key), password)
	_, err := db.Exec(
		`INSERT INTO vault(password,account,username) VALUES($1,$2,$3);`,
		password, account, username)
	Try(err)
}

// SELECT di uno specifico UTENTE
func SelectUser(db *sql.DB, username string) (*User, error) {

	user := User{}
	row := db.QueryRow(`SELECT * FROM users WHERE username=$1`, username)
	err := row.Scan(&user.Username, &user.Password, &user.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		} else {
			panic(err)
		}
	}
	return &user, nil
}

func GetVault(db *sql.DB, username string, key string) []Vault {
	row, err := db.Query(`SELECT id,account,password FROM vault WHERE username=$1`, username)
	defer row.Close()

	vaultArray := make([]Vault, 0)

	for row.Next() {

		var account string
		var password string
		var id int

		err = row.Scan(&id, &account, &password)
		Try(err)
		password, _ = Decrypt([]byte(key), password)
		vaultArray = append(vaultArray, Vault{Id: id, Account: account, Password: password})
	}

	return vaultArray
}

func Try(e error) {
	if e != nil {
		panic(e)
	}
}
