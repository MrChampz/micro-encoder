package repositories_test

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/database"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestJobRepositoryDbInsert(t *testing.T) {
	db := database.NewTestDatabase()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	videoRepo := repositories.NewVideoRepository(db)
	videoRepo.Insert(video)

	job, err := domain.NewJob("output_path", "Pending", video)
	require.Nil(t, err)

	jobRepo := repositories.NewJobRepository(db)
	jobRepo.Insert(job)

	j, err := jobRepo.Find(job.ID)
	require.Nil(t, err)
	require.NotEmpty(t, j.ID)
	require.Equal(t, j.ID, job.ID)
	require.Equal(t, j.VideoID, video.ID)
}

func TestJobRepositoryDbUpdate(t *testing.T) {
	db := database.NewTestDatabase()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	videoRepo := repositories.NewVideoRepository(db)
	videoRepo.Insert(video)

	job, err := domain.NewJob("output_path", "Pending", video)
	require.Nil(t, err)

	jobRepo := repositories.NewJobRepository(db)
	jobRepo.Insert(job)

	job.Status = "Complete"
	jobRepo.Update(job)

	j, err := jobRepo.Find(job.ID)
	require.Nil(t, err)
	require.NotEmpty(t, j.ID)
	require.Equal(t, j.Status, job.Status)
}
