package api

import (
	"context"
	"log"

	"github.com/AbderraoufKhorchani/file-saver/file-service/data"
)

// FileService implements the gRPC FileServiceServer interface
type FileService struct {
	UnimplementedFileServiceServer
}

func (s *FileService) UploadFile(ctx context.Context, req *FileRequest) (*FileResponse, error) {
	// Implement file upload logic

	// Assuming you have a storage service for handling file storage
	file := &data.File{
		UserID:      req.GetUserId(),
		FileName:    req.GetFileName(),
		FileContent: req.GetFileContent(),
		Size:        int64(len(req.GetFileContent())),
		FileType:    req.GetFileType(),
	}

	err := SaveFile(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return nil, err
	}

	// Return a response with appropriate fields
	return &FileResponse{
		FileName: "ahla",
	}, nil
}

func SaveFile(file *data.File) error {

	err := data.SaveFile(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return err
	}

	// Return a response with appropriate fields
	return nil
}

func (s *FileService) GetFile(ctx context.Context, req *FileRequest) (*FileResponse, error) {
	// Assuming you have a storage service for handling file storage
	file, err := GetFile(req.GetUserId(), req.GetFileName())
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return file, err

}

func GetFile(id, name string) (*FileResponse, error) {
	// Assuming you have a storage service for handling file storage
	file, err := data.GetFile(id, name)
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	// Return a response with retrieved file content and name
	return &FileResponse{
		FileContent: file.FileContent,
		FileName:    file.FileName,
		FileType:    file.FileType,
	}, nil
}
