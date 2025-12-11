package ginx

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
)

func DecodeRequest(ctx *gin.Context, req any) error {
	inMap := make(map[string]any)

	// path parameters. e.g. /user/:id
	params := ctx.Params
	for _, param := range params {
		inMap[param.Key] = param.Value
	}

	// query parameters. e.g. /user?id=123
	queries := ctx.Request.URL.Query()
	for k, v := range queries {
		if len(v) > 0 {
			inMap[k] = v[0]
		}
	}

	// body parameters
	if err := parseBody(ctx, inMap); err != nil {
		return err
	}

	// decode to request struct
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           req,
		TagName:          "json", // protobuf use "json" tag in default
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(inMap); err != nil {
		return err
	}

	return nil
}

func parseBody(ctx *gin.Context, inMap map[string]any) error {
	contentType := ctx.ContentType()
	if contentType == "" || ctx.Request.ContentLength <= 0 {
		return nil
	}

	// read the body content
	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	if err = ctx.Request.Body.Close(); err != nil {
		return err
	}

	// restore the io.ReadCloser to its original state
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// parse body according to content type
	var body map[string]any
	switch {
	case strings.Contains(contentType, "application/json"):
		fallthrough
	default:
		// try to parse as json
		body, err = parseJsonBody(rawBody)
	}

	if err != nil {
		return err
	}

	maps.Copy(inMap, body)
	return nil
}

func parseJsonBody(rawBody []byte) (map[string]any, error) {
	var body map[string]any
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return nil, err
	}

	return body, nil
}
