package api

import (
	"bytes"
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/AbderraoufKhorchani/file-saver/broker-service/utils"
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
	File          FilePayload         `json:"file,omitempty"`
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

type FilePayload struct {
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

func MainHandler(c *gin.Context) {

	data := c.PostForm("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action field is required"})
		return
	}
	var requestPayload RequestPayload
	if err := json.Unmarshal([]byte(data), &requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error parsing JSON: %s", err)})
		return
	}

	switch requestPayload.Action {
	case "auth":
		login(c, requestPayload.Auth)
	case "signup":
		signup(c, requestPayload.Register)
	case "checkingAuth":
		CheckToken(c)
	case "uploadFile":
		UploadFile(c, requestPayload.File)
	case "getFile":
		GetFile(c, requestPayload.File)
	case "getAllFilesNames":
		GetAllFilesName(c)
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

	c.JSON(http.StatusOK, jsonFromService.Data)
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

	c.JSON(http.StatusOK, jsonFromService.Data)
}

func CheckToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	token, err := utils.ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusAccepted, "authorized")
}

func UploadFile(c *gin.Context, entry FilePayload) {
	// Extract the JWT token from the request header
	tokenString := c.GetHeader("Authorization")

	// Verify the JWT token
	token, err := utils.ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form"})
		return
	}

	// Retrieve the file from the form data
	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	file := files[0]

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error opening uploaded file"})
		return
	}
	defer src.Close()

	// Read the file contents into a byte slice
	fileContents, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file contents"})
		return
	}

	payload, err := utils.DecodeJWT(tokenString)
	if err != nil {
		// Handle error
		fmt.Println("Error decoding JWT:", err)
		return
	}

	userID := payload["sub"].(string)

	conn, err := grpc.Dial("file-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to File Saving service: %v", err)
	}
	defer conn.Close()

	fileClient := NewFileServiceClient(conn)

	// Token is valid, proceed with file saving

	// Extract file data from the request (you may need to handle file uploads properly)
	fileRequest := AddRequest{
		UserId:      userID,         // replace with actual user ID from the token or your authentication system
		FileName:    entry.FileName, // replace with the actual file name
		FileType:    entry.FileType, // replace with the actual file type
		FileContent: fileContents,   // replace with the actual file content
	}

	fileResponse, err := fileClient.UploadFile(context.Background(), &fileRequest)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	message := "File saved successfully " + fileResponse.FileName

	// TODO: Implement your file saving logic here
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func GetFile(c *gin.Context, entry FilePayload) {

	tokenString := c.GetHeader("Authorization")

	token, err := utils.ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := utils.DecodeJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error decoding JWT": err})
		return
	}

	userID := payload["sub"].(string)

	conn, err := grpc.Dial("file-service:50051", grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to connect to File Saving service": err})
		return
	}
	defer conn.Close()

	fileClient := NewFileServiceClient(conn)

	fileRequest := GetRequest{
		UserId:   userID,
		FileName: entry.FileName,
	}

	fileResponse, err := fileClient.GetFile(context.Background(), &fileRequest)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to get file": err})
		return
	}

	c.JSON(http.StatusOK, fileResponse)
}

func GetAllFilesName(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	token, err := utils.ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := utils.DecodeJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error decoding JWT": err})
		return
	}

	userID := payload["sub"].(string)

	conn, err := grpc.Dial("file-service:50051", grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to connect to File Saving service": err})
		return
	}
	defer conn.Close()

	fileClient := NewFileServiceClient(conn)
	fileRequest := GetRequest{
		UserId: userID,
	}

	allFilesResponse, err := fileClient.GetAllFiles(context.Background(), &fileRequest)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to get file": err})
		return
	}

	c.JSON(http.StatusOK, allFilesResponse.Files)
}
