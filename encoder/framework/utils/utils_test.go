package utils_test

import (
	"encoder/framework/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `
		{
			"id": "1ce009e8-3ccd-46d9-9839-4e3b347f75fc",
			"file_path": "convite.mp4",
			"status": "pending"
		}
	`

	err := utils.IsJson(json)
	require.Nil(t, err)

	json = `something that is not a json`
	err = utils.IsJson(json)
	require.NotNil(t, err)
}
