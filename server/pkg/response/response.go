package response

import (
	"net/http"

	"github.com/chatagent/server/pkg/errcode"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.OK.Int(),
		Message: errcode.OK.Message(),
		Data:    data,
	})
}

func Page(c *gin.Context, total int64, list interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.OK.Int(),
		Message: errcode.OK.Message(),
		Data: PageData{
			Total: total,
			List:  list,
		},
	})
}

func Error(c *gin.Context, code errcode.ErrCode) {
	c.JSON(http.StatusOK, Response{
		Code:    code.Int(),
		Message: code.Message(),
	})
}

func ErrorMsg(c *gin.Context, code errcode.ErrCode, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code.Int(),
		Message: msg,
	})
}
