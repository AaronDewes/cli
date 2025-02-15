package list

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/supabase/cli/internal/testing/apitest"
	"gopkg.in/h2non/gock.v1"
)

func TestOrganizationListCommand(t *testing.T) {
	t.Run("lists all organizations", func(t *testing.T) {
		// Setup in-memory fs
		fsys := afero.NewMemMapFs()
		// Setup valid access token
		token := apitest.RandomAccessToken(t)
		t.Setenv("SUPABASE_ACCESS_TOKEN", string(token))
		// Flush pending mocks after test execution
		defer gock.Off()
		gock.New("https://api.supabase.io").
			Get("/v1/organizations").
			Reply(200).
			JSON([]Organization{
				{
					Id:   "combined-fuchsia-lion",
					Name: "Test Organization",
				},
			})
		// Run test
		assert.NoError(t, Run(fsys))
	})

	t.Run("throws error on failure to load token", func(t *testing.T) {
		assert.Error(t, Run(afero.NewMemMapFs()))
	})

	t.Run("throws error on network error", func(t *testing.T) {
		// Setup in-memory fs
		fsys := afero.NewMemMapFs()
		// Setup valid access token
		token := apitest.RandomAccessToken(t)
		t.Setenv("SUPABASE_ACCESS_TOKEN", string(token))
		// Flush pending mocks after test execution
		defer gock.Off()
		gock.New("https://api.supabase.io").
			Get("/v1/organizations").
			ReplyError(errors.New("network error"))
		// Run test
		assert.Error(t, Run(fsys))
	})

	t.Run("throws error on server unavailable", func(t *testing.T) {
		// Setup in-memory fs
		fsys := afero.NewMemMapFs()
		// Setup valid access token
		token := apitest.RandomAccessToken(t)
		t.Setenv("SUPABASE_ACCESS_TOKEN", string(token))
		// Flush pending mocks after test execution
		defer gock.Off()
		gock.New("https://api.supabase.io").
			Get("/v1/organizations").
			Reply(500).
			JSON(map[string]string{"message": "unavailable"})
		// Run test
		assert.Error(t, Run(fsys))
	})

	t.Run("throws error on malformed json", func(t *testing.T) {
		// Setup in-memory fs
		fsys := afero.NewMemMapFs()
		// Setup valid access token
		token := apitest.RandomAccessToken(t)
		t.Setenv("SUPABASE_ACCESS_TOKEN", string(token))
		// Flush pending mocks after test execution
		defer gock.Off()
		gock.New("https://api.supabase.io").
			Get("/v1/organizations").
			Reply(200).
			JSON(map[string]string{})
		// Run test
		assert.Error(t, Run(fsys))
	})
}
