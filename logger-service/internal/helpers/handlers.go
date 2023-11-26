package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func WriteLog(c *gin.Context) {
	// Read JSON into a variable
	var requestPayload JSONPayload
	if err := readJSON(c, &requestPayload); err != nil {
		errorJSON(c, err)
		return
	}

	// Insert data
	event := LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	if err := Insert(event); err != nil {
		errorJSON(c, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: requestPayload.Data,
	}

	writeJSON(c, http.StatusOK, resp)
}

func GetAll(c *gin.Context) {

	all, err := All()
	if err != nil {
		errorJSON(c, err)
		return
	}

	writeJSON(c, http.StatusOK, all)

}

func GetOne(c *gin.Context) {

	id := c.Param("id")

	logItem, err := GetOneDB(id)
	if err != nil {
		errorJSON(c, err)
		return
	}

	writeJSON(c, http.StatusOK, logItem)

}

func UpdateOne(c *gin.Context) {

	id := c.Param("id")

	// Read JSON into a variable
	var requestPayload JSONPayload
	if err := readJSON(c, &requestPayload); err != nil {
		errorJSON(c, err)
		return
	}

	// Insert data
	event := LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	logItem, err := Update(id, event.Name, event.Data)
	if err != nil {
		errorJSON(c, err)
		return
	}

	writeJSON(c, http.StatusOK, logItem)

}

func DeleteOne(c *gin.Context) {

	id := c.Param("id")

	logItem, err := DeleteOneDB(id)
	if err != nil {
		errorJSON(c, err)
		return
	}

	writeJSON(c, http.StatusOK, logItem)

}
