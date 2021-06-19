package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (service *JobService) Start() error {
	err := service.changeJobStatus("DOWNLOADING")
	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Download(os.Getenv("INPUT_BUCKET_NAME"))
	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("FRAGMENTING")
	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Fragment()
	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("ENCODING")
	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Encode()
	if err != nil {
		return service.failJob(err)
	}

	err = service.performUpload()
	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("FINISHING")
	if err != nil {
		return service.failJob(err)
	}

	err = service.VideoService.Finish()
	if err != nil {
		return service.failJob(err)
	}

	err = service.changeJobStatus("COMPLETED")
	if err != nil {
		return service.failJob(err)
	}

	return nil
}

func (service *JobService) performUpload() error {
	err := service.changeJobStatus("UPLOADING")
	if err != nil {
		return err
	}

	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))
	doneUpload := make(chan string)

	upload := NewVideoUpload()
	upload.OutputBucket = os.Getenv("OUTPUT_BUCKET_NAME")
	upload.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + service.VideoService.Video.ID

	go upload.ProcessUpload(concurrency, doneUpload)
	result := <-doneUpload

	if result != "upload completed" {
		return errors.New(result)
	}

	return err
}

func (service *JobService) changeJobStatus(status string) error {
	var err error

	service.Job.Status = status

	service.Job, err = service.JobRepository.Update(service.Job)
	if err != nil {
		return service.failJob(err)
	}

	return nil
}

func (service *JobService) failJob(error error) error {
	service.Job.Status = "FAILED"
	service.Job.Error = error.Error()

	_, err := service.JobRepository.Update(service.Job)
	if err != nil {
		return err
	}

	return error
}
