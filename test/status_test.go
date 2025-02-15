package integration

// Basic imports
import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	clicmd "github.com/supabase/cli/cmd"
)

type StatusTestSuite struct {
	suite.Suite
	tempDir string
	cmd     *cobra.Command

	params []gin.Params
	mtx    sync.RWMutex
}

// test functions
func (suite *StatusTestSuite) TestStatus() {
	// run command
	status, _, err := suite.cmd.Find([]string{"status"})
	require.NoError(suite.T(), err)

	// set stdout to write into file so we can capture cmd output
	tmpfile, err := ioutil.TempFile(suite.tempDir, "output")
	require.NoError(suite.T(), err)
	defer os.Remove(tmpfile.Name()) // clean up
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	os.Stdout = tmpfile

	// run command
	require.NoError(suite.T(), status.RunE(status, []string{}))

	// check request details
	suite.mtx.RLock()
	defer suite.mtx.RUnlock()
	require.Contains(suite.T(), suite.params, gin.Params{
		gin.Param{
			Key:   "id",
			Value: "supabase_db_" + filepath.Base(suite.tempDir),
		},
	})

	contents, err := ioutil.ReadFile(tmpfile.Name())
	require.NoError(suite.T(), err)
	require.Contains(suite.T(), string(contents), "API URL: http://localhost:54321")
	require.Contains(suite.T(), string(contents), "DB URL: postgresql://postgres:postgres@localhost:54322/postgres")
	require.Contains(suite.T(), string(contents), "Studio URL: http://localhost:54323")
	require.Contains(suite.T(), string(contents), "Inbucket URL: http://localhost:54324")
	require.Contains(suite.T(), string(contents), "anon key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	require.Contains(suite.T(), string(contents), "service_role key: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
}

// hooks
func (suite *StatusTestSuite) SetupTest() {
	// init cli
	suite.cmd = clicmd.GetRootCmd()
	suite.tempDir = NewTempDir(Logger, TempDir)

	// init supabase
	init, _, err := suite.cmd.Find([]string{"init"})
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), init.RunE(init, []string{}))

	// implement mocks
	DockerMock.ContainerInspectHandler = func(c *gin.Context) {
		suite.addParams(c.Copy())
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (suite *StatusTestSuite) TeardownTest() {
	require.NoError(suite.T(), os.Chdir(TempDir))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}

// helper functions
func (suite *StatusTestSuite) addParams(c *gin.Context) {
	suite.mtx.Lock()
	defer suite.mtx.Unlock()
	suite.params = append(suite.params, c.Params)
}
