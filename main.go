package main

import (
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Env struct {
	users interface {
		Get(tgID int) (*User, error)
		GetOrInsert(user *User) error
		List() ([]User, error)
		Insert(user *User) error
		UpdateInfo(user *User, updateUserData *User) error
		SetAdminStatus(tgID int, isAdmin bool) error
		HandlerGetUsers(w http.ResponseWriter, r *http.Request)
		HandlerGetUser(w http.ResponseWriter, r *http.Request)
	}

	ipChecks interface {
		List() ([]IPCheck, error)
		ListByTgID(tgID int, uniq bool) ([]IPCheck, error)
		Insert(ipCheck *IPCheck) error
		Delete(ipCheckID int) error
		HandlerGetHistory(w http.ResponseWriter, r *http.Request)
		HandlerDeleteHistoryRecord(w http.ResponseWriter, r *http.Request)
	}

	errLogs interface{
		Write(p []byte) (n int, err error)
	}
}

func main() {
	// Load dotenv
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connecting to DB
	time.Sleep(5*time.Second)

	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"), os.Getenv("DB_TIMEZONE"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error initializing db session")
	}

	// DB migration
	err = db.AutoMigrate(User{}, IPCheck{}, ErrLog{})
	if err != nil {
		log.Fatal("Error run db migration")
	}

	// Init first admin
	initAdminID, err := strconv.Atoi(os.Getenv("ADMIN_TG_ID"))
	if err != nil {
		log.Fatal("Error parsing ADMIN_TG_ID value from .env file")
	}
	db.Create(&User{TgID: initAdminID, IsAdmin: true})

	// Env
	env := &Env{
		users: &UserModel{db},
		ipChecks: &IPCheckModel{db},
		errLogs:  &ErrLogModel{db},
	}

	// Setup logging
	log.SetLevel(log.ErrorLevel)
	log.SetFormatter(&log.TextFormatter{})
	// TODO remove after testing
	mw := io.MultiWriter(env.errLogs, os.Stdout)
	log.SetOutput(mw)
	//log.SetOutput(DBctx)

	// tg-bot up
	go tgBot(env)

	// web-server up
	API(env)
}