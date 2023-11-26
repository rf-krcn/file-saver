package helpers

import (
	"context"
	"log"

	fl "github.com/AbderraoufKhorchani/file-saver/file-service/pkg/file"
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

func (s *FileService) UploadFile(ctx context.Context, req *fl.AddRequest) (*fl.AddResponse, error) {

	file := &File{
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

	return &fl.AddResponse{
		FileName: f.FileName,
	}, nil
}

func SaveFile(file *File) (*File, error) {

	f, err := SaveFileDB(file)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return nil, err
	}

	return f, nil
}

func (s *FileService) GetFile(ctx context.Context, req *fl.GetRequest) (*fl.GetResponse, error) {

	file, err := GetFile(req.GetUserId(), req.GetFileName())
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return file, err

}

func GetFile(id, name string) (*fl.GetResponse, error) {

	file, err := GetFileDB(id, name)
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		return nil, err
	}

	return &fl.GetResponse{
		FileContent:   file.FileContent,
		FileName:      file.FileName,
		FileType:      file.FileType,
		FileSizeBytes: uint64(file.Size),
		CreatingTime:  file.CreatedAt.String(),
	}, nil
}
