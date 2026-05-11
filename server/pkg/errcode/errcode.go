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

	ContactNotFound       ErrCode = 20006
	ConversationNotFound  ErrCode = 20007
	MessageNotFound       ErrCode = 20008
	CannedResponseNotFound ErrCode = 20009
	ConversationClosed    ErrCode = 20010
	ContactBlocked        ErrCode = 20011
	DuplicateContactInbox ErrCode = 20012
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

	ContactNotFound:        "contact not found",
	ConversationNotFound:   "conversation not found",
	MessageNotFound:        "message not found",
	CannedResponseNotFound: "canned response not found",
	ConversationClosed:     "conversation is closed",
	ContactBlocked:         "contact is blocked",
	DuplicateContactInbox:  "duplicate contact inbox",
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
