package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

func (upload *VideoUpload) UploadObject(
	objectPath string,
	client *s3.Client,
	ctx context.Context,
) error {
	path := strings.Split(objectPath, os.Getenv("LOCAL_STORAGE_PATH")+"/")

	file, err := os.Open(objectPath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(upload.OutputBucket),
		Key:    aws.String(path[1]),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   file,
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	return nil
}

func (upload *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	inChan := make(chan int, runtime.NumCPU())
	returnChan := make(chan string)

	err := upload.loadPaths()
	if err != nil {
		return err
	}

	client, ctx, err := getClientUpload()
	if err != nil {
		return err
	}

	for process := 0; process < concurrency; process++ {
		go upload.uploadWorker(inChan, returnChan, client, ctx)
	}

	go func() {
		for x := 0; x < len(upload.Paths); x++ {
			inChan <- x
		}
		close(inChan)
	}()

	for r := range returnChan {
		if r != "" {
			doneUpload <- r
			break
		}
	}

	return nil
}

func (upload *VideoUpload) uploadWorker(
	inChan chan int,
	returnChan chan string,
	client *s3.Client,
	ctx context.Context,
) {
	for x := range inChan {
		err := upload.UploadObject(upload.Paths[x], client, ctx)
		if err != nil {
			upload.Errors = append(upload.Errors, upload.Paths[x])
			log.Printf("Error during the upload: %v. Error: %v", upload.Paths[x], err)
			returnChan <- err.Error()
		}

		returnChan <- ""
	}

	returnChan <- "upload completed"
}

func (upload *VideoUpload) loadPaths() error {
	err := filepath.Walk(upload.VideoPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			upload.Paths = append(upload.Paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func getClientUpload() (*s3.Client, context.Context, error) {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := s3.NewFromConfig(cfg)

	return client, ctx, nil
}
