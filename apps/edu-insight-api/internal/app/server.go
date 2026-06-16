package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
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

	classroomService, scoreService, dbConn, err := bootstrapServices(cfg)
	if err != nil {
		return nil, err
	}

	httpapi.RegisterRoutes(engine, cfg, classroomService, scoreService)

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

func bootstrapServices(cfg config.Config) (appclassroom.Service, appscore.Service, *sql.DB, error) {
	if cfg.DBDSN == "" {
		log.Println("DB_DSN is empty; using in-memory services")
		return appclassroom.NewMemoryService(), appscore.NewMemoryService(), nil, nil
	}

	conn, err := platformdb.Open(cfg.DBDSN)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := platformdb.Migrate(conn, cfg.MigrationsDir); err != nil {
		conn.Close()
		return nil, nil, nil, err
	}
	classroomService, err := appclassroom.NewPersistentService(appclassroom.NewMySQLRepository(conn))
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}
	scoreService, err := appscore.NewPersistentService(appscore.NewMySQLRepository(conn))
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}
	log.Println("using MySQL services")
	return classroomService, scoreService, conn, nil
}
