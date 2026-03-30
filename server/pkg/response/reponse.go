package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`    // 业务状态码 (0:成功, 其他:失败)
	Data    interface{} `json:"data"`    // 数据载体
	Message string      `json:"message"` // 提示信息
}

// Success 成功返回
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

// Fail 失败返回
func Fail(c *gin.Context, httpCode int, businessCode int, msg string) {
	c.JSON(httpCode, Response{
		Code:    businessCode,
		Data:    nil,
		Message: msg,
	})
}
