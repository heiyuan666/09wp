package response

import "github.com/gin-gonic/gin"

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    data,
	})
}

func OKPage(c *gin.Context, list interface{}, total int64) {
	OK(c, PageResult{List: list, Total: total})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}

