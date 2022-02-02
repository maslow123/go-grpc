package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	v1 "github.com/maslow123/go-grpc/pkg/api/v1"
	"github.com/maslow123/go-grpc/pkg/logger"
	"github.com/maslow123/go-grpc/pkg/protocol/grpc"
	"github.com/maslow123/go-grpc/pkg/protocol/rest"
)

// Config is configuration for Server
type Config struct {
	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string

	// DB DataStore parameters section
	// DatastoreDBHost is host of database
	DatastoreDBHost string
	// DatastoreDBUser string
	DatastoreDBUser string
	// DatastoreDBPassword string
	DatastoreDBPassword string
	// DatastoreDBSchema string
	DatastoreDBSchema string

	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel      int
	LogTimeFormat string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	// get configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.HTTPPort, "http-port", "", "HTTP port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "Database Host")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "", "Database User")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "", "Database Password")
	flag.StringVar(&cfg.DatastoreDBSchema, "db-schema", "", "Database Schema")
	flag.IntVar(&cfg.LogLevel, "log-level", 0, "Global log level")
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", "", "Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")

	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	if len(cfg.HTTPPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", cfg.HTTPPort)
	}

	// Initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {
		return fmt.Errorf("Failed to initialize logger: %v", err)
	}

	// add MySQL driver
	param := "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DatastoreDBUser,
		cfg.DatastoreDBPassword,
		cfg.DatastoreDBHost,
		cfg.DatastoreDBSchema,
		param,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewTodoServiceServer(db)

	// run HTTP gateway
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
