package thirdparty

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type Firestore struct {
	client *firebase.App
	bucket string
	ctx    context.Context
}

func NewFirebaseApp(ctx context.Context, firebaseCredentialFile, bucket string) (*Firestore, error) {
	opt := option.WithCredentialsFile(firebaseCredentialFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	return &Firestore{client: app, bucket: bucket, ctx: ctx}, nil
}

func (f *Firestore) UploadFileToFirebaseStorage(file []byte, folder string, fileName string) (string, string, error) {
	client, err := f.client.Storage(f.ctx)
	if err != nil {
		return "", "", fmt.Errorf("error getting Storage client: %v", err)
	}

	var objString = folder + "/" + fileName
	fmt.Println("BUCKET", f.bucket)
	fmt.Println("objString", objString)
	bucket, err := client.Bucket(f.bucket)
	if err != nil {
		return "", "", fmt.Errorf("error getting bucket: %v", err)
	}

	wc := bucket.Object(objString).NewWriter(f.ctx)
	if _, err = wc.Write(file); err != nil {
		return "", "", fmt.Errorf("error writing object to bucket: %v", err)
	}

	if err := wc.Close(); err != nil {
		return "", "", fmt.Errorf("error closing writer: %v", err)
	}

	fmt.Println("ERROR", f.makePublic(objString))

	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", f.bucket, objString)
	fmt.Printf("Public URL: %s\n", publicURL)

	return objString, publicURL, nil
}

func (f *Firestore) makePublic(object string) error {

	client, err := f.client.Storage(f.ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}

	bucket, err := client.Bucket(f.bucket)
	if err != nil {
		return fmt.Errorf("error getting bucket: %v", err)
	}

	acl := bucket.Object(object).ACL()
	if err := acl.Set(f.ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("ACLHandle.Set: %w", err)
	}
	return nil
}
