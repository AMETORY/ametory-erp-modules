package file

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
)

type FileService struct {
	ctx     *context.ERPContext
	baseURL string
}

// NewFileService creates a new instance of FileService.
//
// It takes an ERPContext and a baseURL as parameter and returns a pointer to a FileService.
//
// It uses the ERPContext to initialize the FileService and run the migration for the FileModel database schema.
func NewFileService(ctx *context.ERPContext, baseURL string) *FileService {
	service := FileService{
		ctx:     ctx,
		baseURL: baseURL,
	}

	if ctx.SkipMigration {
		return &service
	}
	err := ctx.DB.AutoMigrate(&models.FileModel{})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

// UploadFileFromBase64 uploads a file from a base64 encoded string.
//
// It takes a base64 encoded string, a provider, a folder and a pointer to a FileModel as parameter.
//
// It decodes the base64 string into a byte array and then calls UploadFile with the byte array and the given parameters.
//
// It returns an error if there is an error decoding the base64 string or uploading the file.
func (s *FileService) UploadFileFromBase64(base64String, provider, folder string, fileObj *models.FileModel) error {
	file, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Errorf("error decoding base64 string: %v", err)
	}
	return s.UploadFile(file, provider, folder, fileObj)
}

// UploadFile uploads a file to a provider.
//
// It takes a byte array, a provider, a folder and a pointer to a FileModel as parameter.
//
// It detects the mime type of the file from the byte array and then calls the UploadFileToFirebaseStorage method of the Firestore service if the provider is "firebase", or writes the file to the local file system if the provider is "local".
//
// It returns an error if there is an error uploading the file to the provider.
func (s *FileService) UploadFile(file []byte, provider, folder string, fileObj *models.FileModel) error {
	// TODO: implement upload file logic
	var path, url, mimeType string
	mimeType = http.DetectContentType(file)
	fileObj.MimeType = mimeType

	ext := ""
	fileNameSplit := strings.Split(fileObj.FileName, ".")
	if len(fileNameSplit) == 1 {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[0]
		}
		fileObj.FileName = fmt.Sprintf("%s-%d%s", fileObj.FileName, time.Now().UnixMilli(), ext)
	} else {
		fileObj.FileName = fmt.Sprintf("%s-%d.%s", fileNameSplit[0], time.Now().UnixMilli(), fileNameSplit[1])
	}
	switch provider {
	case "local":
		filePath := fmt.Sprintf("./assets/%s/%s", folder, fileObj.FileName)
		err := os.MkdirAll(fmt.Sprintf("./assets/%s", folder), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}

		err = os.WriteFile(filePath, file, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error writing file: %v", err)
		}

		path = filePath
		url = fmt.Sprintf("%s/assets/%s/%s", s.baseURL, folder, fileObj.FileName)

	case "firebase":
		firestoreSrv, ok := s.ctx.Firestore.(*thirdparty.Firestore)
		if !ok {
			return fmt.Errorf("firestore service is not found")
		}
		pathStr, urlStr, err := firestoreSrv.UploadFileToFirebaseStorage(file, folder, fileObj.FileName)
		if err != nil {
			return err
		}
		path = pathStr
		url = urlStr
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	fileObj.Path = path
	fileObj.Provider = "firebase"
	fileObj.URL = url
	fileObj.Provider = provider

	if fileObj.SkipSave {
		return nil
	}

	return s.ctx.DB.Save(fileObj).Error
}

// GetFileByID retrieves a file by its ID.
//
// It takes an ID as parameter and returns the FileModel if found, otherwise an error.
func (s *FileService) GetFileByID(id string) (*models.FileModel, error) {
	file := &models.FileModel{}
	err := s.ctx.DB.Where("id = ?", id).First(file).Error
	if err != nil {
		return nil, err
	}
	return file, nil
}

// UpdateFileByID updates the details of a file in the database by its ID.
//
// It takes a string id and a pointer to a FileModel containing the updated file information.
// The function returns an error if the update operation fails.

func (s *FileService) UpdateFileByID(id string, file *models.FileModel) error {
	return s.ctx.DB.Model(&models.FileModel{}).Where("id = ?", id).Updates(file).Error
}

// UpdateFileRefByID updates the reference ID and type of a file in the database by its ID.
//
// It takes a string id, a string refID, and a string refType as parameters.
// The function returns an error if the update operation fails.
func (s *FileService) UpdateFileRefByID(id string, refID, refType string) error {
	return s.ctx.DB.Model(&models.FileModel{}).Where("id = ?", id).Updates(map[string]interface{}{"ref_id": refID, "ref_type": refType}).Error
}

// DeleteFile deletes a file by its ID.
//
// It takes an ID as parameter and retrieves the associated FileModel from the database.
// If the file provider is "local", it deletes the file from the local file system.
// If the file provider is "firebase", it deletes the file from the Firebase Storage.
// The function then deletes the FileModel from the database.
// It returns an error if any of the operations fail.
func (s *FileService) DeleteFile(id string) error {
	file := &models.FileModel{}
	err := s.ctx.DB.Where("id = ?", id).First(file).Error
	if err != nil {
		return err
	}

	if file.Provider == "local" {
		os.Remove(file.Path)
	}
	if file.Provider == "firebase" {
		firestoreSrv, ok := s.ctx.Firestore.(*thirdparty.Firestore)
		if !ok {
			return fmt.Errorf("firestore service is not found")
		}
		firestoreSrv.DeleteFileFromFirebaseStorage(file.Path)
	}

	return s.ctx.DB.Delete(file).Error
}
