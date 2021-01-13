package upyun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Error struct {
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	RequestID  string `json:"id"`
	Operation  string
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (e *Error) Error() string {
	if e.Operation == "" {
		e.Operation = "upyun api"
	}

	return fmt.Sprintf("%s error: status=%d, code=%d, message=%s, request-id=%s",
		e.Operation, e.StatusCode, e.Code, e.Message, e.RequestID)
}

func checkResponse(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	uerr := new(Error)
	uerr.StatusCode = res.StatusCode
	uerr.Header = res.Header
	defer res.Body.Close()
	slurp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return uerr
	}
	uerr.Body = slurp
	json.Unmarshal(slurp, uerr)
	return uerr
}

func checkStatusCode(err error, status int) bool {
	if err == nil {
		return false
	}

	ae, ok := err.(*Error)
	return ok && ae.StatusCode == status
}

func IsNotExist(err error) bool {
	return checkStatusCode(err, http.StatusNotFound)
}

func IsNotModified(err error) bool {
	return checkStatusCode(err, http.StatusNotModified)
}

func IsTooManyRequests(err error) bool {
	return checkStatusCode(err, http.StatusTooManyRequests)
}

func errorOperation(op string, err error) error {
	if err == nil {
		return errors.New(op)
	}
	ae, ok := err.(*Error)
	if ok {
		ae.Operation = op
		return ae
	}
	return fmt.Errorf("%s: %w", op, err)
}
