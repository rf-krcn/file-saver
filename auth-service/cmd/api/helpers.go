package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"` // Use `interface{}` instead of `any`
}

func readJSON(c *gin.Context, data interface{}) error {
	maxBytes := int64(1048576) // One megabyte

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(data); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

func writeJSON(c *gin.Context, status int, data interface{}, headers ...map[string]string) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			c.Header(key, value)
		}
	}

	c.Header("Content-Type", "application/json")
	c.String(status, string(out))
	return nil
}

func errorJSON(c *gin.Context, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	payload := jsonResponse{
		Error:   true,
		Message: err.Error(),
	}

	return writeJSON(c, statusCode, payload)
}
