package data

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var db *gorm.DB

func New(dbPool *gorm.DB) {
	db = dbPool
	db.AutoMigrate(&File{})
}

type File struct {
	gorm.Model
	UserID      string `gorm:"not null"`
	FileName    string `gorm:"not null"`
	FileContent []byte `gorm:"not null"`
	Size        int64  `gorm:"not null"`
	FileType    string `gorm:"not null"`
}

func SaveFile(file *File) (*File, error) {
	// Check if a file with the same name already exists for the user
	existingFile, err := GetFile(file.UserID, file.FileName)
	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error if there is an issue other than the record not found
		return nil, err
	}

	// If a file with the same name exists, modify the filename
	if existingFile != nil {
		file.FileName = findUniqueFilename(file.FileName, file.UserID)
	}

	// Save the modified or original file to the database
	result := db.Create(file)
	return file, result.Error
}

func GetFile(userID string, fileName string) (*File, error) {
	var file File
	result := db.Where("user_id = ? AND file_name = ?", userID, fileName).First(&file)
	if result.Error != nil {
		return nil, result.Error
	}
	return &file, nil
}

func GetAllFiles(userID string) ([]File, error) {
	var files []File
	result := db.Where("user_id = ?", userID).Find(&files)
	if result.Error != nil {
		return nil, result.Error
	}
	return files, nil

}

func findUniqueFilename(filename, userID string) string {
	// Split the filename and extension
	parts := strings.Split(filename, ".")
	baseName := parts[0]
	extension := ""
	if len(parts) > 1 {
		extension = parts[1]
	}

	// Iterate until a unique filename is found
	i := 1
	for {
		// Form the new filename
		newFilename := fmt.Sprintf("%s(%d)", baseName, i)
		if extension != "" {
			newFilename += fmt.Sprintf(".%s", extension)
		}

		// Check if the file with the new filename exists
		existingFile, _ := GetFile(userID, newFilename)
		if existingFile == nil {
			return newFilename
		}

		i++
	}
}
