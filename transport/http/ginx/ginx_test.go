package ginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type UserRequest struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestDecode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, err := http.NewRequest("POST", "/user/123", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "123"}}

	var userReq UserRequest
	err = DecodeRequest(c, &userReq)
	assert.NoError(t, err)

	assert.Equal(t, 123, userReq.ID)
}
