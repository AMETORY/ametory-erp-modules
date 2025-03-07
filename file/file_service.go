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

func (s *FileService) UploadFileFromBase64(base64String, provider, folder string, fileObj *models.FileModel) error {
	file, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Errorf("error decoding base64 string: %v", err)
	}
	return s.UploadFile(file, provider, folder, fileObj)
}

func (s *FileService) UploadFile(file []byte, provider, folder string, fileObj *models.FileModel) error {
	// TODO: implement upload file logic
	firestoreSrv, ok := s.ctx.Firestore.(*thirdparty.Firestore)
	if !ok && provider != "local" {
		return fmt.Errorf("firestore service is not found")
	}
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

func (s *FileService) UpdateFileByID(id string, file *models.FileModel) error {
	return s.ctx.DB.Model(&models.FileModel{}).Where("id = ?", id).Updates(file).Error
}
func (s *FileService) UpdateFileRefByID(id string, refID, refType string) error {
	return s.ctx.DB.Model(&models.FileModel{}).Where("id = ?", id).Updates(map[string]interface{}{"ref_id": refID, "ref_type": refType}).Error
}
