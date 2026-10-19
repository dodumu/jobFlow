package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ProfilePictureDirectory() (string, error) {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	directory := filepath.Join(uploadPath, "profile-pictures")

	err := os.MkdirAll(directory, 0755)
	if err != nil {
		return "", err
	}

	return directory, nil
}

func GenerateProfilePictureFilename(userID int, contentType string) (string, error) {
	var extension string

	switch strings.ToLower(contentType) {
	case "image/jpeg":
		extension = ".jpg"

	case "image/png":
		extension = ".png"

	case "image/webp":
		extension = ".webp"

	default:
		return "", fmt.Errorf("unsupported profile picture type: %s", contentType)
	}

	filename := fmt.Sprintf(
		"%d-%d%s",
		userID,
		time.Now().UnixNano(),
		extension,
	)

	return filename, nil
}

func SaveProfilePicture(
	file multipart.File,
	userID int,
	contentType string,
) (string, error) {
	directory, err := ProfilePictureDirectory()
	if err != nil {
		return "", fmt.Errorf("creating profile picture directory: %w", err)
	}

	filename, err := GenerateProfilePictureFilename(userID, contentType)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(directory, filename)

	destination, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("creating profile picture file: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, file)
	if err != nil {
		return "", fmt.Errorf("saving profile picture: %w", err)
	}

	return filename, nil
}

func DeleteProfilePicture(filename string) error {
	if filename == "" {
		return nil
	}
	if filepath.Base(filename) != filename || filename == "." || filename == ".." {
		return fmt.Errorf("invalid profile picture filename")
	}

	directory, err := ProfilePictureDirectory()
	if err != nil {
		return fmt.Errorf("locating profile picture directory: %w", err)
	}
	filePath := filepath.Join(directory, filename)

	err = os.Remove(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("deleting profile picture: %w", err)
	}
	return nil
}
