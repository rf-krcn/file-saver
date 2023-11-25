package api

import (
	"bytes"
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	grpc "google.golang.org/grpc"
)

type RequestPayload struct {
	Action        string              `json:"action"`
	Auth          LoginPayload        `json:"auth,omitempty"`
	LogID         string              `json:"log_id,omitempty"`
	UserID        string              `json:"user_id,omitempty"`
	ResetPassword ResetPasswordPaylod `json:"reset_password,omitempty"`
	Register      UserPayload         `json:"signup,omitempty"`
}

type ResetPasswordPaylod struct {
	Email       string `json:"user_name"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

type Config struct{}

type LoginPayload struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type UserPayload struct {
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Password  string `json:"password"`
}

func MainHandler(c *gin.Context) {
	var requestPayload RequestPayload
	if err := c.ShouldBindJSON(&requestPayload); err != nil {
		fmt.Println("error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch requestPayload.Action {
	case "auth":
		login(c, requestPayload.Auth)
	case "signup":
		signup(c, requestPayload.Register)
	}
}

func login(c *gin.Context, entry LoginPayload) {

	authServiceURL := "http://auth-service/login/"

	jsonData, err := json.Marshal(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	request, err := http.NewRequest("POST", authServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.StatusCode})
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if jsonFromService.Error {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	c.JSON(http.StatusAccepted, jsonFromService.Data)
}

func signup(c *gin.Context, entry UserPayload) {

	jsonData, err := json.Marshal(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	authServiceURL := "http://auth-service/signup"

	request, err := http.NewRequest("POST", authServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create an HTTP client and make the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Status})
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if jsonFromService.Error {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Signing in failed"})
		return
	}

	c.JSON(http.StatusAccepted, jsonFromService.Data)

}

func UploadFileHandler(c *gin.Context) {
	// Extract the JWT token from the request header
	/*
		tokenString := c.GetHeader("Authorization")

		// Verify the JWT token
		token, err := utils.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	*/

	conn, err := grpc.Dial("file-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to File Saving service: %v", err)
	}
	defer conn.Close()

	fileClient := NewFileServiceClient(conn)

	// Token is valid, proceed with file saving

	// Extract file data from the request (you may need to handle file uploads properly)
	fileRequest := FileRequest{
		UserId:      "user123",                 // replace with actual user ID from the token or your authentication system
		FileName:    "example.txt",             // replace with the actual file name
		FileType:    "text/plain",              // replace with the actual file type
		FileContent: []byte("example content"), // replace with the actual file content
	}

	_, err = fileClient.UploadFile(context.Background(), &fileRequest)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	// TODO: Implement your file saving logic here
	c.JSON(http.StatusOK, gin.H{"message": "File saved successfully"})
}
