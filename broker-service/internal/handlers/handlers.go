package handlers

import (
	"bytes"
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AbderraoufKhorchani/file-saver/broker-service/pkg/file"
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
	FileSize int    `json:"file_size"`
}
type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message,omitempty"`
	Token   interface{} `json:"token,omitempty"`
	User    interface{} `json:"user,omitempty"`
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
	case "getAllFiles":
		GetAllFilesName(c)
	}

}

func login(c *gin.Context, entry LoginPayload) {

	authServiceURL := "http://auth-service/login/"

	jsonData, err := json.Marshal(entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error parsing JSON: %s", err)})
		return
	}

	request, err := http.NewRequest("POST", authServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Internal server error: %s", err)})
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Internal server error: %s", err)})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusOK {
		c.JSON(response.StatusCode, gin.H{"error": response.Body})
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

	c.JSON(http.StatusOK, jsonFromService)
}

func signup(c *gin.Context, entry UserPayload) {

	jsonData, err := json.Marshal(entry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		c.JSON(http.StatusBadRequest, jsonFromService.Message)
		return
	}

	if jsonFromService.Error {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Signing in failed"})
		return
	}

	c.JSON(http.StatusOK, jsonFromService)
}

func CheckToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	fmt.Println(tokenString)

	token, err := ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusAccepted, "authorized")
}

func UploadFile(c *gin.Context, entry FilePayload) {
	tokenString := c.GetHeader("Authorization")
	token, err := ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := DecodeJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding JWT"})
		return
	}

	userID := payload["sub"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "file-service:50051", grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to File Saving service"})
		return
	}
	defer conn.Close()

	fileClient := file.NewFileServiceClient(conn)
	stream, err := fileClient.UploadFile(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error opening gRPC stream"})
		return
	}
	defer stream.CloseSend()

	fileRequest := &file.AddRequest{
		UserId:   userID,
		FileName: entry.FileName,
		FileType: entry.FileType,
		FileSize: int64(entry.FileSize),
	}

	if err := stream.Send(fileRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending metadata to gRPC stream"})
		return
	}

	fileTemp, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting file from request"})
		return
	}
	defer fileTemp.Close()

	bufferSize := 4096
	buffer := make([]byte, bufferSize)

	for {
		n, err := fileTemp.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
			return
		}

		chunk := &file.AddRequest{
			FileContent: buffer[:n],
		}
		if err := stream.Send(chunk); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending file content to gRPC stream"})
			return
		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error receiving response from gRPC stream"})
		return
	}

	message := "File saved successfully"

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func GetFile(c *gin.Context, entry FilePayload) {
	tokenString := c.GetHeader("Authorization")

	token, err := ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := DecodeJWT(tokenString)
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

	fileClient := file.NewFileServiceClient(conn)

	fileRequest := &file.GetRequest{
		UserId:   userID,
		FileName: entry.FileName,
	}

	// Open a gRPC stream to receive file information and content
	stream, err := fileClient.GetFile(context.Background(), fileRequest)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to get file": err})
		return
	}
	defer stream.CloseSend()

	// Create a buffer to accumulate file content chunks
	c.Header("Content-Type", "application/octet-stream")

	// Set other headers as needed
	c.Header("Content-Disposition", "attachment; filename="+entry.FileName)

	// Create a buffer to accumulate file content chunks
	//var fileContent []byte

	for {
		contentChunk, err := stream.Recv()
		if err == io.EOF {
			// End of file streaming
			break
		}
		if err != nil {
			// Handle error
			return
		}

		// Append the content chunk to the buffer
		//fileContent = append(fileContent, contentChunk.FileContent...)

		// Stream the content chunk to the client
		c.Writer.Write(contentChunk.FileContent)
		c.Writer.Flush()
	}

}

func GetAllFilesName(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	token, err := ValidateToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := DecodeJWT(tokenString)
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

	fileClient := file.NewFileServiceClient(conn)
	fileRequest := file.GetRequest{
		UserId: userID,
	}

	allFilesResponse, err := fileClient.GetAllFiles(context.Background(), &fileRequest)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Failed to get file": err})
		return
	}

	err = logRequest("file checking", fmt.Sprintf("%s checked his files", userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allFilesResponse.Files)
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

//promonom
