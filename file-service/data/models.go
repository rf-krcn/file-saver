package data

import "gorm.io/gorm"

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

func SaveFile(file *File) error {
	result := db.Create(file)
	return result.Error
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
	result := db.Where("user_id = ?", userID).First(&files)
	if result.Error != nil {
		return nil, result.Error
	}

	return files, nil

}
