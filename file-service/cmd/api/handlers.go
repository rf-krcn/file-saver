package api

import (
	"context"
	"log"

	"github.com/AbderraoufKhorchani/file-saver/file-service/data"
)

type FileService struct {
	UnimplementedFileServiceServer
}

func (s *FileService) GetAllFiles(ctx context.Context, req *GetRequest) (*AllFilesResponse, error) {

	response, err := GetAllFiles(req.UserId)
	if err != nil {
		log.Printf("Error getting files: %v", err)
		return nil, err
	}

	return response, nil
}

func GetAllFiles(userID string) (*AllFilesResponse, error) {

	files, err := data.GetAllFiles(userID)
	if err != nil {
		log.Printf("Error getting files: %v", err)
		return nil, err
	}

	var responses []*GetResponse
	for _, file := range files {
		responses = append(responses, &GetResponse{
			FileName:      file.FileName,
			FileType:      file.FileType,
			FileSizeBytes: uint64(file.Size),
			CreatingTime:  file.CreatedAt.String(),
		})
	}

	allFilesResponse := &AllFilesResponse{
		Files: responses,
	}

	return allFilesResponse, nil
}

func (s *FileService) UploadFile(ctx context.Context, req *AddRequest) (*AddResponse, error) {

	file := &data.File{
		UserID:      req.GetUserId(),
		FileName:    req.GetFileName(),
		FileContent: req.GetFileContent(),
		Size:        int64(len(req.GetFileContent())),
		FileType:    req.GetFileType(),
	}

	f, err := SaveFile(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return nil, err
	}

	return &AddResponse{
		FileName: f.FileName,
	}, nil
}

func SaveFile(file *data.File) (*data.File, error) {

	f, err := data.SaveFile(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return nil, err
	}

	return f, nil
}

func (s *FileService) GetFile(ctx context.Context, req *GetRequest) (*GetResponse, error) {

	file, err := GetFile(req.GetUserId(), req.GetFileName())
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return file, err

}

func GetFile(id, name string) (*GetResponse, error) {

	file, err := data.GetFile(id, name)
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return &GetResponse{
		FileContent:   file.FileContent,
		FileName:      file.FileName,
		FileType:      file.FileType,
		FileSizeBytes: uint64(file.Size),
		CreatingTime:  file.CreatedAt.String(),
	}, nil
}
