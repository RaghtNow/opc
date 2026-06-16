package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/httpapi"
	platformdb "github.com/RaghtNow/opc/apps/edu-insight-api/internal/platform/db"
)

type Server struct {
	engine *gin.Engine
	cfg    config.Config
	db     *sql.DB
}

func NewServer() (*Server, error) {
	cfg := config.Load()

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	classroomService, dbConn, err := bootstrapClassroomService(cfg)
	if err != nil {
		return nil, err
	}

	httpapi.RegisterRoutes(engine, cfg, classroomService)

	return &Server{
		engine: engine,
		cfg:    cfg,
		db:     dbConn,
	}, nil
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return http.ListenAndServe(addr, s.engine)
}

func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func bootstrapClassroomService(cfg config.Config) (appclassroom.Service, *sql.DB, error) {
	if cfg.DBDSN == "" {
		log.Println("DB_DSN is empty; using in-memory classroom service")
		return appclassroom.NewMemoryService(), nil, nil
	}

	conn, err := platformdb.Open(cfg.DBDSN)
	if err != nil {
		return nil, nil, err
	}
	if err := platformdb.Migrate(conn, cfg.MigrationsDir); err != nil {
		conn.Close()
		return nil, nil, err
	}
	service, err := appclassroom.NewPersistentService(appclassroom.NewMySQLRepository(conn))
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	log.Println("using MySQL classroom service")
	return service, conn, nil
}
