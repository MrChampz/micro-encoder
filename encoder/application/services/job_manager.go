package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/queue"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db               *gorm.DB
	Job              domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManager(
	db *gorm.DB,
	rabbitMQ *queue.RabbitMQ,
	messageChannel chan amqp.Delivery,
	jobReturnChannel chan JobWorkerResult,
) *JobManager {
	return &JobManager{
		Db:               db,
		Job:              domain.Job{},
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (manager *JobManager) Start(channel *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.NewVideoRepository(manager.Db)

	jobService := JobService{
		JobRepository: repositories.NewJobRepository(manager.Db),
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))

	if err != nil {
		log.Fatalf("error loading var:")
	}

	for processes := 0; processes < concurrency; processes++ {
		go JobWorker(
			manager.MessageChannel,
			manager.JobReturnChannel,
			jobService,
			manager.Job,
			processes,
		)
	}

	for jobResult := range manager.JobReturnChannel {
		if jobResult.Error != nil {
			err = manager.checkParseErrors(jobResult)
		} else {
			err = manager.notifySuccess(jobResult, channel)
		}

		if err != nil {
			jobResult.Message.Reject(false)
		}
	}
}

func (manager *JobManager) notifySuccess(jobResult JobWorkerResult, channel *amqp.Channel) error {
	Mutex.Lock()
	jobJson, err := json.Marshal(jobResult.Job)
	Mutex.Unlock()

	if err != nil {
		return err
	}

	err = manager.notify(jobJson)
	if err != nil {
		return err
	}

	err = jobResult.Message.Ack(false)
	if err != nil {
		return err
	}

	return nil
}

func (manager *JobManager) checkParseErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Printf(
			"MessageID: %v. Error during the job: %v, with video: %v. Error: %v.",
			jobResult.Message.DeliveryTag, jobResult.Job.ID, jobResult.Job.Video.ID, jobResult.Error.Error(),
		)
	} else {
		log.Printf(
			"MessageID: %v. Error parsing message: %v.",
			jobResult.Message.DeliveryTag, jobResult.Error.Error(),
		)
	}

	jobError := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error:   jobResult.Error.Error(),
	}

	errorJson, err := json.Marshal(jobError)
	if err != nil {
		return err
	}

	err = manager.notify(errorJson)
	if err != nil {
		return err
	}

	err = jobResult.Message.Reject(false)
	if err != nil {
		return err
	}

	return nil
}

func (manager *JobManager) notify(json []byte) error {
	err := manager.RabbitMQ.Notify(
		string(json),                                   // message
		"application/json",                             // content-type
		os.Getenv("RABBITMQ_NOTIFICATION_EX"),          // exchange
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"), // routing key
	)

	if err != nil {
		return err
	}

	return nil
}
