package helpers

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	fl "github.com/AbderraoufKhorchani/file-saver/file-service/pkg/file"
	"gorm.io/gorm"
)

type FileService struct {
	fl.UnimplementedFileServiceServer
}

func (s *FileService) GetAllFiles(ctx context.Context, req *fl.GetRequest) (*fl.AllFilesResponse, error) {

	response, err := GetAllFiles(req.UserId)
	if err != nil {
		log.Printf("Error getting files: %v", err)
		return nil, err
	}

	return response, nil
}

func GetAllFiles(userID string) (*fl.AllFilesResponse, error) {

	files, err := GetAllFilesDB(userID)
	if err != nil {
		log.Printf("Error getting files: %v", err)
		return nil, err
	}

	var responses []*fl.GetResponse
	for _, file := range files {
		responses = append(responses, &fl.GetResponse{
			FileName:      file.FileName,
			FileType:      file.FileType,
			FileSizeBytes: uint64(file.Size),
			CreatingTime:  file.CreatedAt.String(),
		})
	}

	allFilesResponse := &fl.AllFilesResponse{
		Files: responses,
	}

	return allFilesResponse, nil
}

func (s *FileService) UploadFile(stream fl.FileService_UploadFileServer) error {
	var metadata *fl.AddRequest

	if metadataChunk, err := stream.Recv(); err == nil {
		metadata = metadataChunk
	} else {
		return err
	}

	userID := metadata.GetUserId()
	fileName := metadata.GetFileName()
	fileType := metadata.GetFileType()
	fileSize := metadata.GetFileSize()

	existingFile, err := GetFileDB(userID, fileName)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existingFile != nil {
		fileName = findUniqueFilename(fileName, userID)
	}

	fileStruct := &File{
		UserID:   userID,
		FileName: fileName,
		FileType: fileType,
		Size:     fileSize,
	}

	// Specify the directory where you want to store the files
	storageDirectory := "/files"

	// Construct the directory path based on user ID
	userDirectory := filepath.Join(storageDirectory, userID)

	// Ensure the user-specific directory exists
	if err := os.MkdirAll(userDirectory, os.ModePerm); err != nil {
		log.Printf("Error creating user directory: %v", err)
		return err
	}

	// Construct the full path to the file inside the user's directory
	fullFilePath := filepath.Join(userDirectory, fileName)

	// Create or open the file for writing
	file, err := os.Create(fullFilePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return err
	}
	defer file.Close()
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			return err
		}

		// Write the content chunk to the file
		_, err = file.Write(chunk.FileContent)
		if err != nil {
			log.Printf("Error writing to file: %v", err)
			return err
		}
	}

	_, err = SaveFile(fileStruct)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return err
	}

	response := &fl.AddResponse{
		FileName: fileName,
	}

	return stream.SendAndClose(response)
}

func SaveFile(file *File) (*File, error) {

	f, err := SaveFileDB(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return nil, err
	}

	return f, nil
}

func (s *FileService) GetFile(req *fl.GetRequest, stream fl.FileService_GetFileServer) error {

	filePath := filepath.Join("/files", req.GetUserId(), req.GetFileName())
	fileReader, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer fileReader.Close()

	// Stream the file content to the client
	const chunkSize = 4096
	buffer := make([]byte, chunkSize)
	for {
		bytesRead, err := fileReader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading file: %v", err)
			return err
		}

		// Send the chunk to the client
		if err := stream.Send(&fl.GetContent{FileContent: buffer[:bytesRead]}); err != nil {
			log.Printf("Error sending file content: %v", err)
			return err
		}
	}

	return nil
}

func GetFile(id, name string) (*fl.GetResponse, error) {

	file, err := GetFileDB(id, name)
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return &fl.GetResponse{
		FileName:      file.FileName,
		FileType:      file.FileType,
		FileSizeBytes: uint64(file.Size),
		CreatingTime:  file.CreatedAt.String(),
	}, nil
}
