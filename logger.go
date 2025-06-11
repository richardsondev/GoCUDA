package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelSuccess LogLevel = "success"
	LogLevelDebug   LogLevel = "debug"
	LogLevelCUDA    LogLevel = "cuda"
)

// LogMessage represents a log entry
type LogMessage struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
	Source    string   `json:"source"`
}

// Logger handles WebSocket connections and broadcasts log messages
type Logger struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan LogMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	history    []LogMessage
	maxHistory int
}

// Global logger instance
var globalLogger *Logger

// Initialize the logger
func initLogger() {
	globalLogger = &Logger{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan LogMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		history:    make([]LogMessage, 0),
		maxHistory: 100,
	}

	go globalLogger.run()

	// Initial system info
	LogInfo("System", "Logger initialized")
	LogInfo("System", fmt.Sprintf("Operating System: Windows"))
	LogInfo("System", fmt.Sprintf("Application: CUDA Torus Visualization"))
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocket handler
func wsLogHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	globalLogger.register <- conn

	// Send log history to new client
	globalLogger.mutex.RLock()
	for _, msg := range globalLogger.history {
		if err := conn.WriteJSON(msg); err != nil {
			globalLogger.mutex.RUnlock()
			globalLogger.unregister <- conn
			conn.Close()
			return
		}
	}
	globalLogger.mutex.RUnlock()

	// Keep connection alive and handle disconnection
	defer func() {
		globalLogger.unregister <- conn
		conn.Close()
	}()

	// Read messages from client (ping/pong)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Run the logger hub
func (l *Logger) run() {
	for {
		select {
		case client := <-l.register:
			l.mutex.Lock()
			l.clients[client] = true
			l.mutex.Unlock()

			LogInfo("WebSocket", "New client connected to log stream")

		case client := <-l.unregister:
			l.mutex.Lock()
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				client.Close()
				LogInfo("WebSocket", "Client disconnected from log stream")
			}
			l.mutex.Unlock()

		case message := <-l.broadcast:
			// Add to history
			l.mutex.Lock()
			l.history = append(l.history, message)
			if len(l.history) > l.maxHistory {
				l.history = l.history[1:]
			}
			l.mutex.Unlock()

			// Broadcast to all clients
			l.mutex.RLock()
			for client := range l.clients {
				err := client.WriteJSON(message)
				if err != nil {
					delete(l.clients, client)
					client.Close()
				}
			}
			l.mutex.RUnlock()
		}
	}
}

// LogInfo logs an info message
func LogInfo(source, message string) {
	logMessage(LogLevelInfo, source, message)
	fmt.Printf("[INFO] %s: %s\n", source, message)
}

// LogWarning logs a warning message
func LogWarning(source, message string) {
	logMessage(LogLevelWarning, source, message)
	fmt.Printf("[WARNING] %s: %s\n", source, message)
}

// LogError logs an error message
func LogError(source, message string) {
	logMessage(LogLevelError, source, message)
	fmt.Printf("[ERROR] %s: %s\n", source, message)
}

// LogSuccess logs a success message
func LogSuccess(source, message string) {
	logMessage(LogLevelSuccess, source, message)
	fmt.Printf("[SUCCESS] %s: %s\n", source, message)
}

// LogDebug logs a debug message
func LogDebug(source, message string) {
	logMessage(LogLevelDebug, source, message)
	fmt.Printf("[DEBUG] %s: %s\n", source, message)
}

// LogCUDA logs a CUDA-specific message
func LogCUDA(source, message string) {
	logMessage(LogLevelCUDA, source, message)
	fmt.Printf("[CUDA] %s: %s\n", source, message)
}

// Helper function to send log messages
func logMessage(level LogLevel, source, message string) {
	if globalLogger == nil {
		return
	}

	logMsg := LogMessage{
		Timestamp: time.Now().Format("15:04:05.000"),
		Level:     level,
		Message:   message,
		Source:    source,
	}

	select {
	case globalLogger.broadcast <- logMsg:
	default:
		// Channel is full, drop the message
	}
}
