package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedFileTypes = map[string]bool{
	"image/jpeg":                                      true,
	"image/png":                                       true,
	"image/gif":                                       true,
	"image/webp":                                      true,
	"application/pdf":                                 true,
	"application/msword":                              true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":                        true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"text/plain":                                      true,
	"text/csv":                                        true,
}

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorMsg(c, errcode.InvalidParam, "no file uploaded")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !allowedFileTypes[contentType] {
		response.ErrorMsg(c, errcode.InvalidParam, "file type not allowed")
		return
	}

	maxSize := int64(20 * 1024 * 1024) // 20MB default
	if allowedImageTypes[contentType] {
		maxSize = 5 * 1024 * 1024 // 5MB for images
	}
	if header.Size > maxSize {
		response.ErrorMsg(c, errcode.InvalidParam, "file too large")
		return
	}

	dateDir := time.Now().Format("2006-01-02")
	dir := filepath.Join("uploads", dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}

	newName := uuid.New().String() + ext
	destPath := filepath.Join(dir, newName)

	dst, err := os.Create(destPath)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	url := fmt.Sprintf("/%s", strings.ReplaceAll(destPath, "\\", "/"))

	response.Success(c, gin.H{
		"url":       url,
		"name":      header.Filename,
		"size":      header.Size,
		"type":      contentType,
	})
}
