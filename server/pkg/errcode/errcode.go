package errcode

type ErrCode int

const (
	OK           ErrCode = 0
	UnknownError ErrCode = 10000
	InvalidParam ErrCode = 10001
	Unauthorized ErrCode = 10002
	Forbidden    ErrCode = 10003
	NotFound     ErrCode = 10004
	ServerError  ErrCode = 10005
	DBError      ErrCode = 10006
	TokenExpired ErrCode = 10007

	UserNotFound  ErrCode = 20001
	WrongPassword ErrCode = 20002
	UserDisabled  ErrCode = 20003
	EmailExists   ErrCode = 20004
	ContactExists ErrCode = 20005
)

var messages = map[ErrCode]string{
	OK:           "success",
	UnknownError: "unknown error",
	InvalidParam: "invalid parameter",
	Unauthorized: "unauthorized",
	Forbidden:    "forbidden",
	NotFound:     "not found",
	ServerError:  "internal server error",
	DBError:      "database error",
	TokenExpired: "token expired",

	UserNotFound:  "user not found",
	WrongPassword: "wrong password",
	UserDisabled:  "user account is disabled",
	EmailExists:   "email already registered",
	ContactExists: "contact already exists",
}

func (e ErrCode) Message() string {
	if msg, ok := messages[e]; ok {
		return msg
	}
	return messages[UnknownError]
}

func (e ErrCode) Int() int {
	return int(e)
}
