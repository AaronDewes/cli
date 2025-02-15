package supabase

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	IDAlphabet  = "abcdefghijklmnopqrstuvwxyz"
	IDLength    = 20
	KeyAlphabet = "abcdef0123456789"
	KeyLength   = 40
)

// Server struct with route handlers
type Server struct {
	FunctionsHandler func(c *gin.Context)
	SecretsHandler   func(c *gin.Context)
}

var defaultHandler = func(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Not implemented",
	})
}

// NewServer creates a new server with default handlers
func NewServer() *Server {
	s := Server{
		FunctionsHandler: defaultHandler,
		SecretsHandler:   defaultHandler,
	}
	return &s
}

// NewRouter creating a new router and setting the routes for the server.
func (s *Server) NewRouter() *gin.Engine {
	root := gin.Default()
	router := root.Group("/v1")

	projects := router.Group("/projects")
	projects.GET("/:id/functions", s.functions)
	projects.GET("/:id/secrets", s.secrets)

	return root
}

// project routes
func (s *Server) functions(c *gin.Context) {
	if s.FunctionsHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "handler is nil",
		})
	} else {
		s.FunctionsHandler(c)
	}
}

func (s *Server) secrets(c *gin.Context) {
	if s.SecretsHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "handler is nil",
		})
	} else {
		s.SecretsHandler(c)
	}
}
