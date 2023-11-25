package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AbderraoufKhorchani/file-saver/auth-service/data"
	"github.com/AbderraoufKhorchani/file-saver/auth-service/utils"
	"github.com/gin-gonic/gin"
)

type Config struct {
}

func Register(c *gin.Context) {

	userJSON := data.UserJSONBinding{}

	if err := c.ShouldBindJSON(&userJSON); err != nil {
		errorJSON(c, err)
		return
	}

	user := data.User{

		UserName:  userJSON.UserName,
		FirstName: userJSON.FirstName,
		LastName:  userJSON.LastName,
		Password:  userJSON.Password,
	}

	err := data.Insert(user)
	if err != nil {
		errorJSON(c, err)
		return
	}

	/*
		err = logRequest("adding user", fmt.Sprintf("User %s added.", user.UserName))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	*/

	token, err := utils.GenerateToken(user.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Respond with the generated token
	c.JSON(http.StatusAccepted, gin.H{
		"error":   false,
		"message": "Signed up!",
		"data":    token,
	})

}

func Login(c *gin.Context) {
	var loginRequest struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Retrieve the user by username
	user, err := data.GetByUserName(loginRequest.UserName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the provided password with the stored hashed password
	err = utils.ComparePassword(user.Password, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate a JWT token
	token, err := utils.GenerateToken(user.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Respond with the generated token
	c.JSON(http.StatusAccepted, gin.H{
		"error":   false,
		"message": "Authenticated!",
		"data":    token,
	})
}

func (app *Config) ResetPassword(c *gin.Context) {

	var requestPayload struct {
		UserName    string `json:"user_name"`
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&requestPayload); err != nil {
		errorJSON(c, err)
		return
	}

	user, err := data.GetByUserName(requestPayload.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	err = utils.ComparePassword(user.Password, requestPayload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	err = data.ResetPassword(requestPayload.NewPassword, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	err = logRequest("password resetting", fmt.Sprintf("%s password reset ", user.UserName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resonse := user.UserName + "'s Password resetted."

	c.JSON(http.StatusOK, resonse)

}

func logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.Marshal(entry)
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
