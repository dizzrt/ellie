package ginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock protobuf generated struct
type GetServiceProviderRequest struct {
	SpId uint32 `json:"sp_id,omitempty"`
	Base *Base  `json:"base,omitempty"`
}

type Base struct {
	Version string `json:"version,omitempty"`
}

func TestDecodeRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, err := http.NewRequest("GET", "/test?sp_id=10000", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	var protoReq GetServiceProviderRequest
	err = DecodeRequest(c, &protoReq)
	assert.NoError(t, err)

	// verify if sp_id is correctly bound
	assert.Equal(t, uint32(10000), protoReq.SpId)
}
