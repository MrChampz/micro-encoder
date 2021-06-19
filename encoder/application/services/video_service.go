package services

import (
	"bytes"
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (service *VideoService) Download(bucketName string) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)

	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(service.Video.FilePath),
	})
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	_, err = body.ReadFrom(obj.Body)
	if err != nil {
		return err
	}

	defer obj.Body.Close()

	basePath := os.Getenv("LOCAL_STORAGE_PATH") + "/"

	file, err := os.Create(basePath + service.Video.ID + ".mp4")
	if err != nil {
		return err
	}

	_, err = file.Write(body.Bytes())
	if err != nil {
		return err
	}

	defer file.Close()

	log.Printf("Video %v has been stored", service.Video.ID)

	return nil
}

func (service *VideoService) Fragment() error {
	basePath := os.Getenv("LOCAL_STORAGE_PATH") + "/"

	err := os.Mkdir(basePath+service.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}

	source := basePath + service.Video.ID + ".mp4"
	target := basePath + service.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (service *VideoService) Encode() error {
	basePath := os.Getenv("LOCAL_STORAGE_PATH") + "/"

	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, basePath+service.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, basePath+service.Video.ID)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")
	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (service *VideoService) Finish() error {
	basePath := os.Getenv("LOCAL_STORAGE_PATH") + "/"

	err := os.Remove(basePath + service.Video.ID + ".mp4")
	if err != nil {
		log.Println("Error removing mp4 ", service.Video.ID)
		return err
	}

	err = os.Remove(basePath + service.Video.ID + ".frag")
	if err != nil {
		log.Println("Error removing frag ", service.Video.ID)
		return err
	}

	err = os.RemoveAll(basePath + service.Video.ID)
	if err != nil {
		log.Println("Error removing mp4dash content", service.Video.ID)
		return err
	}

	log.Println("Files have been removed ", service.Video.ID)

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("======> Output: %s\n", string(out))
	}
}
