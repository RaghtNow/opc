package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/httpapi"
)

type Server struct {
	engine *gin.Engine
	cfg    config.Config
}

func NewServer() (*Server, error) {
	cfg := config.Load()

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	httpapi.RegisterRoutes(engine, cfg)

	return &Server{
		engine: engine,
		cfg:    cfg,
	}, nil
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return http.ListenAndServe(addr, s.engine)
}
