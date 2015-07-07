package usererror

import (
	"fmt"
	"net/http"
)

type QueryUserError struct {
	Message  string
	HTTPCode int
}

func (ue *QueryUserError) Error() string {
	return ue.Message
}

func UserErrorF(format string, params ...interface{}) *QueryUserError {
	return &QueryUserError{Message: fmt.Sprintf(format, params...), HTTPCode: http.StatusBadRequest}
}

func UserAccessErrorF(format string, params ...interface{}) *QueryUserError {
	return &QueryUserError{Message: fmt.Sprintf(format, params...), HTTPCode: http.StatusForbidden}
}

func (ue *QueryUserError) GetUserDescription() string {
	return ue.Message
}
func (ue *QueryUserError) GetHTTPStatus() int {
	return ue.HTTPCode
}
