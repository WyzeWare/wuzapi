package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"gopkg.in/natefinch/lumberjack.v2"
	_ "modernc.org/sqlite"
)

type server struct {
	db     *sql.DB
	router *mux.Router
	exPath string
}

var (
	// Flags that can be set via command line

	address    = flag.String("address", "0.0.0.0", "Bind IP Address")
	port       = flag.String("port", "8080", "Listen Port")
	waDebug    = flag.String("wadebug", "", "Enable whatsmeow debug (INFO or DEBUG)")
	logType    = flag.String("logtype", "console", "Type of log output (console or json)")
	sslcert    = flag.String("sslcertificate", "", "SSL Certificate File")
	sslprivkey = flag.String("sslprivatekey", "", "SSL Certificate Private Key File")
	adminToken = flag.String("admintoken", "", "Security Token to authorize admin actions (list/create/remove users)")

	configFile  = flag.String("config", "/etc/wuzapi/config", "Path to the configuration file")
	postgresCfg = flag.String("postgresconfig", "/etc/wuzapi/postgres_config", "Path to the PostgreSQL configuration file")

	container     *sqlstore.Container
	killchannel   = make(map[int](chan bool))
	userinfocache = cache.New(5*time.Minute, 10*time.Minute)
	log           zerolog.Logger
)

// Config represents the parsed configuration data
type Config map[string]string

// ParseConfigFile parses the configuration file with the specified filename
func ParseConfigFile(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	config := make(Config)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return config, nil
}

func init() {
	flag.Parse()

	// Set up logging to file
	logPath := "/var/log/wuzapi/wuzapi.log"
	if os.Getenv("WUZAPI_LOG_PATH") != "" {
		logPath = os.Getenv("WUZAPI_LOG_PATH")
	}

	logFile := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	}

	log = zerolog.New(logFile).With().Timestamp().Str("role", filepath.Base(os.Args[0])).Logger()
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get executable path")
	}
	exPath := filepath.Dir(ex)

	config, err := ParseConfigFile(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read configuration file")
	}

	dbType := config["DB_TYPE"]
	var db *sql.DB

	switch dbType {
	case "sqlite3":
		dbPath := config["DB_PATH"]
		db, err = sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on")
		if err != nil {
			log.Fatal().Err(err).Msg("Could not open SQLite database")
		}

		if *waDebug == "" {
			dbLog := waLog.Stdout("Database", *waDebug, true)
			container, err = sqlstore.New("sqlite3", "file:"+dbPath+"?_foreign_keys=on&_busy_timeout=3000", dbLog)
		} else {
			container, err = sqlstore.New("sqlite3", "file:"+dbPath+"?_foreign_keys=on&_busy_timeout=3000", nil)
		}

	case "postgresql":
		pgConfig, err := ParseConfigFile(*postgresCfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read PostgreSQL configuration file")
		}
		connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			pgConfig["HOST"], pgConfig["USER"], pgConfig["PASSWORD"], pgConfig["DATABASE"])
		db, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not open PostgreSQL database")
		}

		if *waDebug == "" {
			dbLog := waLog.Stdout("Database", *waDebug, true)
			container, err = sqlstore.New("postgres", connectionString, dbLog)
		} else {
			container, err = sqlstore.New("postgres", connectionString, nil)
		}

	default:
		log.Fatal().Msg("Invalid database type specified")
	}
	defer db.Close()

	s := &server{
		router: mux.NewRouter(),
		db:     db,
		exPath: exPath,
	}
	s.routes()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var srv *http.Server

	if *sslcert != "" && *sslprivkey != "" {
		srv = &http.Server{
			Addr:    *address + ":" + *port,
			Handler: s.router,
		}
		go func() {
			if err := srv.ListenAndServeTLS(*sslcert, *sslprivkey); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Server startup failed")
			}
		}()
	} else {
		srv = &http.Server{
			Addr:    *address + ":" + *port,
			Handler: s.router,
		}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Server startup failed")
			}
		}()
	}

	log.Info().Str("address", *address).Str("port", *port).Msg("Server started")

	<-done
	log.Info().Msg("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}
	log.Info().Msg("Server exited properly")
}
