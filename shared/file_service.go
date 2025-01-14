package shared

import (
	"fmt"
	"net/http"
	"os"

	"github.com/AMETORY/ametory-erp-modules/context"
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
	err := Migrate(ctx.DB)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *FileService) UploadFile(file []byte, provider, folder string, fileObj *FileModel) error {
	// TODO: implement upload file logic
	firestoreSrv, ok := s.ctx.Firestore.(*thirdparty.Firestore)
	if !ok {
		return fmt.Errorf("firestore service is not found")
	}
	var path, url, mimeType string
	mimeType = http.DetectContentType(file)
	fileObj.MimeType = mimeType

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

	return s.ctx.DB.Save(fileObj).Error
}
