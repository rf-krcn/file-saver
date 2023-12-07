package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
}

func Register(c *gin.Context) {

	userJSON := UserJSONBinding{}

	if err := c.ShouldBindJSON(&userJSON); err != nil {
		errorJSON(c, err)
		return
	}

	user := User{

		UserName:  userJSON.UserName,
		FirstName: userJSON.FirstName,
		LastName:  userJSON.LastName,
		Password:  userJSON.Password,
	}

	err := Insert(user)
	if err != nil {
		errorJSON(c, err)
		return
	}

	err = logRequest("Adding user", fmt.Sprintf("User %s added.", user.UserName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := GenerateToken(user.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""

	// Respond with the generated token
	c.JSON(http.StatusAccepted, gin.H{
		"error":   false,
		"message": "Signed up!",
		"token":   token,
		"user":    user,
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
	user, err := GetByUserName(loginRequest.UserName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the provided password with the stored hashed password
	err = ComparePassword(user.Password, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = logRequest("User authentificated", fmt.Sprintf("User %s logged in.", user.UserName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate a JWT token
	token, err := GenerateToken(user.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""

	// Respond with the generated token
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Authenticated!",
		"token":   token,
		"user":    user,
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

	user, err := GetByUserName(requestPayload.UserName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	err = ComparePassword(user.Password, requestPayload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	err = ResetPassword(requestPayload.NewPassword, user.ID)
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
