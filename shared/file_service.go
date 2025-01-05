package shared

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
)

type FileService struct {
	ctx *context.ERPContext
}

func NewFileService(ctx *context.ERPContext) *FileService {
	service := FileService{
		ctx: ctx,
	}

	err := Migrate(ctx.DB)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *FileService) UploadFile(file []byte, folder string, fileObj FileModel) (*FileModel, error) {
	// TODO: implement upload file logic
	firestoreSrv, ok := s.ctx.Firestore.(*thirdparty.Firestore)
	if !ok {
		return nil, fmt.Errorf("firestore service is not found")
	}
	path, url, err := firestoreSrv.UploadFileToFirebaseStorage(file, folder, fileObj.FileName)
	if err != nil {
		return nil, err
	}
	fileObj.Path = path
	fileObj.Provider = "firebase"
	fileObj.URL = url
	return &fileObj, nil
}
