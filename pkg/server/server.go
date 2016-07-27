package server

import (
	"flag"
	"net/http"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krivokhatko/wschat/pkg/config"
	"github.com/krivokhatko/wschat/pkg/logger"
	"github.com/krivokhatko/wschat/pkg/switcher"
	uuid "github.com/nu7hatch/gouuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	ID         string
	Config     *config.Config
	configPath string
	logPath    string
	logLevel   int
	startTime  time.Time
	stopping   bool
}

func NewServer(configPath, logPath string, logLevel int) *Server {
	return &Server{
		Config: &config.Config{
			BindAddr: config.DefaultBindAddr,
			BaseDir:  config.DefaultBaseDir,
		},
		configPath: configPath,
		logPath:    logPath,
		logLevel:   logLevel,
	}
}

func (server *Server) Run() error {
	server.startTime = time.Now()
	// Set up server logging
	if server.logPath != "" && server.logPath != "-" {
		dirPath, _ := path.Split(server.logPath)
		os.MkdirAll(dirPath, 0755)

		serverOutput, err := os.OpenFile(server.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Log(logger.LevelError, "server", "unable to open log file: %s", err)
			return err
		}

		defer serverOutput.Close()

		logger.SetOutput(serverOutput)
	}

	logger.SetLevel(server.logLevel)

	// Load server configuration
	if err := server.Config.Load(server.configPath); err != nil {
		logger.Log(logger.LevelError, "server", "unable to load configuration: %s", err)
		return err
	}

	// Generate unique server instance identifier
	uuidTemp, err := uuid.NewV4()
	if err != nil {
		return err
	}

	server.ID = uuidTemp.String()

	templatePath := server.Config.BaseDir + "template/home.html"
	homeTemplate := template.Must(template.ParseFiles(templatePath))

	flag.Parse()
	switcher := switcher.NewSwitcher()
	go switcher.Proceed()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHome(homeTemplate, w, r)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(switcher, w, r)
	})
	err = http.ListenAndServe(server.Config.BindAddr, nil)
	if err != nil {
		logger.Log(logger.LevelError, "server", "ListenAndServe: %s", err)
		return err
	}

	return nil
}

// Stop stops the server.
func (server *Server) Stop() {
	if server.stopping {
		return
	}

	logger.Log(logger.LevelNotice, "server", "shutting down server")

	server.stopping = true

}

func serveHome(templ *template.Template, w http.ResponseWriter, r *http.Request) {
	logger.Log(logger.LevelNotice, "server", r.URL.Path)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templ.Execute(w, r.Host)
}

// serveWs handles websocket requests from the peer.
func serveWs(sw *switcher.Switcher, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log(logger.LevelError, "server", "Upgrade: %s", err)
		return
	}
	user := &switcher.User{MsgSwitcher: sw, Conn: conn, Send: make(chan *switcher.ChatMessage, 256)}
	user.MsgSwitcher.Login <- user
	go user.WritePump()
	user.ReadPump()
}
