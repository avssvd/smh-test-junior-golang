package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type Response struct {
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
	User           *User     `json:"user,omitempty"`
	Users          []User    `json:"users,omitempty"`
	IPCheckHistory []IPCheck `json:"ip_check_history,omitempty"`
}

func (resp *Response) toJSON() ([]byte, error) {
	respByte, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return respByte, nil
}

func NewErrorResponse(err error) *Response {
	response := new(Response)
	response.Success = false
	response.Error = err.Error()
	return response
}

func queryCheckMiddleware(r *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/get_user", "/get_history_by_tg":
				err := idCheck(req.URL.Query().Get("userTgID"), "userTgID")
				if err != nil {
					badResp, _ := NewErrorResponse(err).toJSON()
					w.WriteHeader(http.StatusBadRequest)
					_, err = w.Write(badResp)
					if err != nil {
						log.Error(err)
					}
					return
				}

			case "/delete_history_record":
				err := idCheck(req.URL.Query().Get("ipCheckID"), "ipCheckID")
				if err != nil {
					badResp, _ := NewErrorResponse(err).toJSON()
					w.WriteHeader(http.StatusBadRequest)
					_, err = w.Write(badResp)
					if err != nil {
						log.Error(err)
					}
					return
				}
			}

			next.ServeHTTP(w, req)
		})
	}
}

func idCheck(strID string, paramName string) error {
	if len(strID) == 0 {
		return fmt.Errorf("parameter '%v' not found", paramName)
	}

	id, err := strconv.Atoi(strID)
	if err != nil || id < 0 {
		return fmt.Errorf("invalid value for query parameter '%v'. Must be unsigned integer", paramName)
	}

	return nil
}

func API(env *Env) {
	r := mux.NewRouter()
	r.Use(queryCheckMiddleware(r))
	r.HandleFunc("/get_users", env.users.HandlerGetUsers)
	r.HandleFunc("/get_user", env.users.HandlerGetUser)
	r.HandleFunc("/get_history_by_tg", env.ipChecks.HandlerGetHistory)
	r.HandleFunc("/delete_history_record", env.ipChecks.HandlerDeleteHistoryRecord)

	fmt.Println("starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
