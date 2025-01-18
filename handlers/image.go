package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const maxImageSize = 20 * 1024 * 1024 // 20 MB
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func validateImage(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > maxImageSize {
		return errors.New("the image is too large, maximum size is 20 MB")
	}

	// Check file type
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		return errors.New("failed to read the image file")
	}
	fileType := http.DetectContentType(buf)
	if !allowedImageTypes[fileType] {
		return errors.New("unsupported image type, allowed types are JPEG, PNG, and GIF")
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.New("failed to reset file pointer")
	}

	return nil
}

func saveImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Create the uploads directory if it doesn't exist
	uploadsDir := "uploads"
	// if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
	// 	return "", errors.New("failed to create uploads directory")
	// }

	// Generate a unique filename
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	filePath := uploadsDir + "/" + fileName

	// Save the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", errors.New("failed to save the image")
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", errors.New("failed to copy the image")
	}

	return filePath, nil
}
