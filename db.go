package main

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	TgID           int `gorm:"primaryKey"`
	TgUserName     string
	TgFirstName    string
	TgLastName     string
	TgLanguageCode string
	IsAdmin        bool `gorm:"not null;default false"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type IPCheck struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	IP       string
	IPInfo   datatypes.JSON
	UserTgID int
	User     User `json:"-"`

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ErrLog struct {
	ID    int `gorm:"primaryKey;autoIncrement"`
	Error string

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

var ErrUserNotFound = errors.New("user not found")

type UserModel struct {
	DB *gorm.DB
}

func (um *UserModel) Get(tgID int) (*User, error) {
	user := User{}
	if result := um.DB.First(&user, tgID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (um *UserModel) GetOrInsert(user *User) error {
	if result := um.DB.Where(User{TgID: user.TgID}).FirstOrCreate(user); result.Error != nil {
		return result.Error
	}
	return nil
}

func (um *UserModel) List() ([]User, error) {
	users := make([]User, 0, 5)
	if result := um.DB.Find(&users); result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (um *UserModel) Insert(user *User) error {
	if result := um.DB.Create(user); result.Error != nil {
		return result.Error
	}
	return nil
}

func (um *UserModel) UpdateInfo(user *User, updateUserData *User) error {
	updateData := map[string]interface{}{
		"tg_user_name":     updateUserData.TgUserName,
		"tg_first_name":    updateUserData.TgFirstName,
		"tg_last_name":     updateUserData.TgLastName,
		"tg_language_code": updateUserData.TgLanguageCode,
	}
	if result := um.DB.Model(user).Where("tg_id = ?", user.TgID).Updates(updateData); result.Error != nil {
		return result.Error
	}
	return nil
}

func (um UserModel) SetAdminStatus(tgID int, isAdmin bool) error {
	user, err := um.Get(tgID)
	if errors.Is(err, ErrUserNotFound) {
		return err
	}
	if user.IsAdmin == isAdmin {
		return nil
	}

	if result := um.DB.Model(user).Update("is_admin", isAdmin); result.Error != nil {
		return result.Error
	}

	return nil
}

func (um *UserModel) HandlerGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := um.List()
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	resp := Response{
		Success: true,
		Users:   users,
	}

	respByte, err := resp.toJSON()
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respByte)
	if err != nil {
		log.Error(err)
	}
}

func (um *UserModel) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	userTgID, err := strconv.Atoi(r.FormValue("userTgID"))
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	user, err := um.Get(userTgID)
	switch {
	case errors.Is(err, ErrUserNotFound):
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return

	case err != nil:
		log.Error(err)
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	resp := Response{
		Success: true,
		User:    user,
	}

	respByte, err := resp.toJSON()
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respByte)
	if err != nil {
		log.Error(err)
	}
}

type IPCheckModel struct {
	DB *gorm.DB
}

func (ipcm *IPCheckModel) List() ([]IPCheck, error) {
	ipChecks := make([]IPCheck, 0, 5)
	if result := ipcm.DB.Find(&ipChecks); result.Error != nil {
		return nil, result.Error
	}

	return ipChecks, nil
}

func (ipcm *IPCheckModel) ListByTgID(tgID int, uniq bool) ([]IPCheck, error) {
	ipChecks := make([]IPCheck, 0, 5)
	if uniq {
		if result := ipcm.DB.Where("user_tg_id = ?", tgID).Distinct("ip", "ip_info", "user_tg_id").Find(&ipChecks); result.Error != nil {
			return nil, result.Error
		}
	} else {
		if result := ipcm.DB.Where("user_tg_id = ?", tgID).Find(&ipChecks); result.Error != nil {
			return nil, result.Error
		}
	}
	return ipChecks, nil
}

func (ipcm *IPCheckModel) Insert(ipCheck *IPCheck) error {
	err := ipcm.DB.Create(ipCheck).Error
	if err != nil {
		return err
	}
	return nil
}

func (ipcm *IPCheckModel) Delete(ipCheckID int) error {
	if result := ipcm.DB.Delete(&IPCheck{}, ipCheckID); result.Error != nil {
		return result.Error
	}
	return nil
}

func (ipcm *IPCheckModel) HandlerGetHistory(w http.ResponseWriter, r *http.Request) {
	userTgID, err := strconv.Atoi(r.FormValue("userTgID"))
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	ipChecks, err := ipcm.ListByTgID(userTgID, false)
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	resp := Response{
		Success:        true,
		IPCheckHistory: ipChecks,
	}

	respByte, err := resp.toJSON()
	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respByte)
	if err != nil {
		log.Error(err)
	}
}

func (ipcm *IPCheckModel) HandlerDeleteHistoryRecord(w http.ResponseWriter, r *http.Request) {
	userTgID, err := strconv.Atoi(r.FormValue("ipCheckID"))
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	err = ipcm.Delete(userTgID)
	if err != nil {
		badResp, err := NewErrorResponse(err).toJSON()
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(badResp)
		if err != nil {
			log.Error(err)
		}
		return
	}

	resp := Response{
		Success: true,
	}

	respByte, err := resp.toJSON()
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respByte)
	if err != nil {
		log.Error(err)
	}
}

type ErrLogModel struct {
	DB *gorm.DB
}

func (elm *ErrLogModel) Write(p []byte) (n int, err error) {
	errLog := ErrLog{Error: string(p)}
	if result := elm.DB.Create(&errLog); result.Error != nil {
		return 0, err
	}
	return len(errLog.Error), nil
}

func getNewUser(user *tgbotapi.User) *User {
	if user != nil {
		return &User{
			TgID:           user.ID,
			TgUserName:     user.UserName,
			TgFirstName:    user.FirstName,
			TgLastName:     user.LastName,
			TgLanguageCode: user.LanguageCode,
			IsAdmin:        false,
		}
	} else {
		return &User{}
	}
}
