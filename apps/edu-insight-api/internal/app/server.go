package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/auth"
	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appinsight "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/insight"
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

	authService, classroomService, scoreService, insightService, dbConn, err := bootstrapServices(cfg)
	if err != nil {
		return nil, err
	}

	httpapi.RegisterRoutes(engine, cfg, authService, classroomService, scoreService, insightService)

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

func bootstrapServices(cfg config.Config) (appauth.Service, appclassroom.Service, appscore.Service, appinsight.Service, *sql.DB, error) {
	if cfg.DBDSN == "" {
		log.Println("DB_DSN is empty; using in-memory services")
		authService := appauth.NewService(appauth.NewMemoryRepository(), cfg.AppEnv)
		classroomService := appclassroom.NewMemoryService()
		scoreService := appscore.NewMemoryService()
		insightService := appinsight.NewService(classroomService, scoreService, appinsight.NewMemoryRepository())
		return authService, classroomService, scoreService, insightService, nil, nil
	}

	conn, err := platformdb.Open(cfg.DBDSN)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := platformdb.Migrate(conn, cfg.MigrationsDir); err != nil {
		conn.Close()
		return nil, nil, nil, nil, nil, err
	}
	authService := appauth.NewService(appauth.NewMySQLRepository(conn), cfg.AppEnv)
	classroomService, err := appclassroom.NewPersistentService(appclassroom.NewMySQLRepository(conn))
	if err != nil {
		conn.Close()
		return nil, nil, nil, nil, nil, err
	}
	scoreService, err := appscore.NewPersistentService(appscore.NewMySQLRepository(conn))
	if err != nil {
		conn.Close()
		return nil, nil, nil, nil, nil, err
	}
	insightService := appinsight.NewService(classroomService, scoreService, appinsight.NewMySQLRepository(conn))
	log.Println("using MySQL services")
	return authService, classroomService, scoreService, insightService, conn, nil
}
