package domain_test

import (
	"encoder/domain"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateWithValidVideo(t *testing.T) {
	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "1"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	err := video.Validate()
	require.Nil(t, err)
}

func TestValidateWithEmptyVideo(t *testing.T) {
	video := domain.NewVideo()

	err := video.Validate()
	require.Error(t, err)
}

func TestValidateWithVideoIdNotUUID(t *testing.T) {
	video := domain.NewVideo()
	video.ID = "ABC"
	video.ResourceID = "1"
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	err := video.Validate()
	require.Error(t, err)
}
