package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// TorusParams holds the configurable parameters for the torus
type TorusParams struct {
	MajorRadius    float32 `json:"majorRadius"`
	MinorRadius    float32 `json:"minorRadius"`
	MajorSegments  int     `json:"majorSegments"`
	MinorSegments  int     `json:"minorSegments"`
	RotationSpeedX float32 `json:"rotationSpeedX"`
	RotationSpeedY float32 `json:"rotationSpeedY"`
}

// AsteroidParams holds configurable parameters for the asteroid field
type AsteroidParams struct {
	SpawnRate     float32 `json:"spawnRate"`
	MaxAsteroids  int     `json:"maxAsteroids"`
	MinSize       float32 `json:"minSize"`
	MaxSize       float32 `json:"maxSize"`
	MinSpeed      float32 `json:"minSpeed"`
	MaxSpeed      float32 `json:"maxSpeed"`
	RotationSpeed float32 `json:"rotationSpeed"`
}

var currentParams = TorusParams{
	MajorRadius:    2.0,
	MinorRadius:    0.5,
	MajorSegments:  50,
	MinorSegments:  30,
	RotationSpeedX: 0.3,
	RotationSpeedY: 0.5,
}

var currentAsteroidParams = AsteroidParams{
	SpawnRate:     1.0,
	MaxAsteroids:  500,
	MinSize:       0.05,
	MaxSize:       0.3,
	MinSpeed:      0.5,
	MaxSpeed:      3.0,
	RotationSpeed: 1.0,
}

var currentScene = "donut" // "donut" or "asteroids"

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>CUDA-Accelerated 3D Visualization Control Panel</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 0; 
            padding: 20px;
            background: #f0f0f0; 
            display: flex;
            gap: 20px;
            height: 100vh;
            box-sizing: border-box;
        }
        .main-panel { 
            flex: 1; 
            max-width: 500px;
            background: white; 
            padding: 30px; 
            border-radius: 10px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            height: fit-content;
        }
        .log-panel {
            flex: 1;
            background: white;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
            min-height: 0;
        }
        .log-header {
            padding: 20px 20px 10px 20px;
            border-bottom: 1px solid #eee;
            background: #f8f9fa;
            border-radius: 10px 10px 0 0;
        }
        .log-header h2 {
            margin: 0;
            color: #333;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .status-indicator {
            width: 10px;
            height: 10px;
            border-radius: 50%;
            background: #dc3545;
            display: inline-block;
        }
        .status-indicator.connected {
            background: #28a745;
        }
        .log-console {
            flex: 1;
            padding: 10px;
            overflow-y: auto;
            background: #1e1e1e;
            color: #fff;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 12px;
            line-height: 1.4;
            min-height: 0;
        }
        .log-entry {
            margin: 2px 0;
            padding: 2px 0;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        .log-timestamp {
            color: #888;
            margin-right: 8px;
        }
        .log-source {
            color: #ffeb3b;
            margin-right: 8px;
            font-weight: bold;
        }
        .log-level-info { color: #17a2b8; }
        .log-level-success { color: #28a745; }
        .log-level-warning { color: #ffc107; }
        .log-level-error { color: #dc3545; }
        .log-level-debug { color: #6c757d; }
        .log-level-cuda { color: #76b900; }
        
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .param-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        input[type="range"] { width: 100%; margin: 10px 0; }
        .value-display { 
            font-family: monospace; 
            background: #f8f8f8; 
            padding: 5px; 
            border-radius: 3px; 
            display: inline-block; 
            min-width: 60px; 
            text-align: center; 
        }
        .status { text-align: center; margin: 20px 0; padding: 10px; border-radius: 5px; }
        .status.connected { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .status.disconnected { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .info { background: #e9ecef; padding: 15px; border-radius: 5px; margin-top: 20px; }
        button { 
            background: #007bff; 
            color: white; 
            border: none; 
            padding: 10px 20px; 
            border-radius: 5px; 
            cursor: pointer; 
            margin: 5px; 
        }
        button:hover { background: #0056b3; }
        .log-controls {
            padding: 10px 20px;
            border-top: 1px solid #eee;
            background: #f8f9fa;
            border-radius: 0 0 10px 10px;
        }
        .log-controls button {
            font-size: 12px;
            padding: 5px 10px;
            margin: 2px;
        }
        .clear-btn { background: #dc3545 !important; }
        .clear-btn:hover { background: #c82333 !important; }
        
        .scene-selector {
            margin-bottom: 30px;
            text-align: center;
        }
        .scene-button {
            background: #6c757d;
            margin: 0 5px;
            padding: 8px 16px;
        }
        .scene-button.active {
            background: #28a745;
        }
        .scene-button:hover {
            background: #5a6268;
        }
        .scene-button.active:hover {
            background: #218838;
        }
        .scene-panel {
            display: none;
        }
        .scene-panel.active {
            display: block;
        }
    </style>
</head>
<body>
    <div class="main-panel">
        <h1>🚀 CUDA-Accelerated 3D Visualization</h1>
        
        <div class="status connected">
            <strong>Status:</strong> <span id="statusText">3D Renderer Active</span>
        </div>

        <div class="scene-selector">
            <button class="scene-button active" id="donutBtn" onclick="switchScene('donut')">🍩 Torus Scene</button>
            <button class="scene-button" id="asteroidsBtn" onclick="switchScene('asteroids')">🌌 Asteroid Field</button>
        </div>

        <!-- Torus Scene Controls -->
        <div class="scene-panel active" id="donutPanel">
            <h3>🍩 Torus Parameters</h3>
            
            <div class="param-group">
                <label for="majorRadius">Major Radius (Torus Size): <span class="value-display" id="majorRadiusValue">{{.MajorRadius}}</span></label>
                <input type="range" id="majorRadius" min="0.5" max="5.0" step="0.1" value="{{.MajorRadius}}" oninput="updateTorusParam('majorRadius', this.value)">
            </div>

            <div class="param-group">
                <label for="minorRadius">Minor Radius (Tube Thickness): <span class="value-display" id="minorRadiusValue">{{.MinorRadius}}</span></label>
                <input type="range" id="minorRadius" min="0.1" max="2.0" step="0.05" value="{{.MinorRadius}}" oninput="updateTorusParam('minorRadius', this.value)">
            </div>

            <div class="param-group">
                <label for="majorSegments">Major Segments (Quality): <span class="value-display" id="majorSegmentsValue">{{.MajorSegments}}</span></label>
                <input type="range" id="majorSegments" min="8" max="100" step="2" value="{{.MajorSegments}}" oninput="updateTorusParam('majorSegments', this.value)">
            </div>

            <div class="param-group">
                <label for="minorSegments">Minor Segments (Smoothness): <span class="value-display" id="minorSegmentsValue">{{.MinorSegments}}</span></label>
                <input type="range" id="minorSegments" min="6" max="60" step="2" value="{{.MinorSegments}}" oninput="updateTorusParam('minorSegments', this.value)">
            </div>

            <div class="param-group">
                <label for="rotationSpeedX">X Rotation Speed: <span class="value-display" id="rotationSpeedXValue">{{.RotationSpeedX}}</span></label>
                <input type="range" id="rotationSpeedX" min="-2.0" max="2.0" step="0.1" value="{{.RotationSpeedX}}" oninput="updateTorusParam('rotationSpeedX', this.value)">
            </div>

            <div class="param-group">
                <label for="rotationSpeedY">Y Rotation Speed: <span class="value-display" id="rotationSpeedYValue">{{.RotationSpeedY}}</span></label>
                <input type="range" id="rotationSpeedY" min="-2.0" max="2.0" step="0.1" value="{{.RotationSpeedY}}" oninput="updateTorusParam('rotationSpeedY', this.value)">
            </div>

            <div style="text-align: center; margin: 20px 0;">
                <button onclick="resetTorusDefaults()">Reset to Defaults</button>
                <button onclick="randomizeTorus()">Randomize</button>
            </div>
        </div>

        <!-- Asteroid Field Controls -->
        <div class="scene-panel" id="asteroidsPanel">
            <h3>🌌 Asteroid Field Parameters</h3>
            
            <div class="param-group">
                <label for="spawnRate">Spawn Rate (asteroids/sec): <span class="value-display" id="spawnRateValue">{{.SpawnRate}}</span></label>
                <input type="range" id="spawnRate" min="0.1" max="5.0" step="0.1" value="{{.SpawnRate}}" oninput="updateAsteroidParam('spawnRate', this.value)">
            </div>

            <div class="param-group">
                <label for="maxAsteroids">Max Asteroids: <span class="value-display" id="maxAsteroidsValue">{{.MaxAsteroids}}</span></label>
                <input type="range" id="maxAsteroids" min="50" max="1000" step="50" value="{{.MaxAsteroids}}" oninput="updateAsteroidParam('maxAsteroids', this.value)">
            </div>

            <div class="param-group">
                <label for="minSize">Min Size: <span class="value-display" id="minSizeValue">{{.MinSize}}</span></label>
                <input type="range" id="minSize" min="0.01" max="0.2" step="0.01" value="{{.MinSize}}" oninput="updateAsteroidParam('minSize', this.value)">
            </div>

            <div class="param-group">
                <label for="maxSize">Max Size: <span class="value-display" id="maxSizeValue">{{.MaxSize}}</span></label>
                <input type="range" id="maxSize" min="0.1" max="1.0" step="0.05" value="{{.MaxSize}}" oninput="updateAsteroidParam('maxSize', this.value)">
            </div>

            <div class="param-group">
                <label for="minSpeed">Min Speed: <span class="value-display" id="minSpeedValue">{{.MinSpeed}}</span></label>
                <input type="range" id="minSpeed" min="0.1" max="2.0" step="0.1" value="{{.MinSpeed}}" oninput="updateAsteroidParam('minSpeed', this.value)">
            </div>

            <div class="param-group">
                <label for="maxSpeed">Max Speed: <span class="value-display" id="maxSpeedValue">{{.MaxSpeed}}</span></label>
                <input type="range" id="maxSpeed" min="1.0" max="10.0" step="0.5" value="{{.MaxSpeed}}" oninput="updateAsteroidParam('maxSpeed', this.value)">
            </div>

            <div class="param-group">
                <label for="rotationSpeed">Rotation Speed: <span class="value-display" id="rotationSpeedValue">{{.RotationSpeed}}</span></label>
                <input type="range" id="rotationSpeed" min="0.1" max="5.0" step="0.1" value="{{.RotationSpeed}}" oninput="updateAsteroidParam('rotationSpeed', this.value)">
            </div>

            <div style="text-align: center; margin: 20px 0;">
                <button onclick="resetAsteroidDefaults()">Reset to Defaults</button>
                <button onclick="clearAsteroids()">Clear All Asteroids</button>
            </div>
        </div>

        <div class="info">
            <strong>Info:</strong> This control panel adjusts the 3D visualization in real-time. 
            Switch between the torus scene and asteroid field using the buttons above or press 1/2 in the 3D viewport.
            <br><br>
            <strong>Current Mode:</strong> <span id="renderMode">{{if .CudaEnabled}}CUDA Accelerated{{else}}CPU Fallback{{end}}</span>
            <br>
            <strong>Keyboard Shortcuts:</strong> 1 = Torus Scene, 2 = Asteroid Field, ESC = Exit
        </div>
    </div>

    <div class="log-panel">
        <div class="log-header">
            <h2>
                📊 Application Logs
                <span class="status-indicator" id="wsStatus"></span>
            </h2>
        </div>
        <div class="log-console" id="logConsole">
            <div class="log-entry log-level-info">
                <span class="log-timestamp">00:00:00</span>
                <span class="log-source">[System]</span>
                Connecting to log stream...
            </div>
        </div>
        <div class="log-controls">
            <button onclick="clearLogs()" class="clear-btn">Clear Logs</button>
            <button onclick="toggleAutoScroll()" id="autoScrollBtn">Auto-scroll: ON</button>
            <button onclick="exportLogs()">Export Logs</button>
        </div>
    </div>

    <script>
        let ws = null;
        let autoScroll = true;
        let logEntries = [];

        // WebSocket connection for logs
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws/logs';
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                console.log('WebSocket connected');
                document.getElementById('wsStatus').classList.add('connected');
                addLogEntry({
                    timestamp: getCurrentTime(),
                    level: 'success',
                    source: 'WebSocket',
                    message: 'Connected to log stream'
                });
            };
            
            ws.onmessage = function(event) {
                try {
                    const logData = JSON.parse(event.data);
                    addLogEntry(logData);
                } catch (e) {
                    console.error('Failed to parse log message:', e);
                }
            };
            
            ws.onclose = function() {
                console.log('WebSocket disconnected');
                document.getElementById('wsStatus').classList.remove('connected');
                addLogEntry({
                    timestamp: getCurrentTime(),
                    level: 'error',
                    source: 'WebSocket',
                    message: 'Disconnected from log stream'
                });
                
                // Reconnect after 3 seconds
                setTimeout(connectWebSocket, 3000);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
                addLogEntry({
                    timestamp: getCurrentTime(),
                    level: 'error',
                    source: 'WebSocket',
                    message: 'Connection error occurred'
                });
            };
        }

        function addLogEntry(logData) {
            logEntries.push(logData);
            
            const logConsole = document.getElementById('logConsole');
            const logEntry = document.createElement('div');
            logEntry.className = 'log-entry log-level-' + logData.level;
            
            logEntry.innerHTML = 
                '<span class="log-timestamp">' + logData.timestamp + '</span>' +
                '<span class="log-source">[' + logData.source + ']</span>' +
                logData.message;
                
            logConsole.appendChild(logEntry);
            
            // Limit log entries to prevent memory issues
            if (logConsole.children.length > 1000) {
                logConsole.removeChild(logConsole.firstChild);
                logEntries.shift();
            }
            
            if (autoScroll) {
                logConsole.scrollTop = logConsole.scrollHeight;
            }
        }

        function getCurrentTime() {
            const now = new Date();
            return now.getHours().toString().padStart(2, '0') + ':' +
                   now.getMinutes().toString().padStart(2, '0') + ':' +
                   now.getSeconds().toString().padStart(2, '0') + '.' +
                   now.getMilliseconds().toString().padStart(3, '0');
        }

        function clearLogs() {
            document.getElementById('logConsole').innerHTML = '';
            logEntries = [];
        }

        function toggleAutoScroll() {
            autoScroll = !autoScroll;
            document.getElementById('autoScrollBtn').textContent = 'Auto-scroll: ' + (autoScroll ? 'ON' : 'OFF');
        }

        function exportLogs() {
            const logText = logEntries.map(entry => 
                entry.timestamp + ' [' + entry.source + '] ' + entry.message
            ).join('\n');
            
            const blob = new Blob([logText], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'torus-app-logs-' + new Date().toISOString().slice(0,19).replace(/:/g, '-') + '.txt';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        }

        function switchScene(scene) {
            // Update UI
            document.querySelectorAll('.scene-button').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('.scene-panel').forEach(panel => panel.classList.remove('active'));
            
            document.getElementById(scene + 'Btn').classList.add('active');
            document.getElementById(scene + 'Panel').classList.add('active');
            
            // Update status text
            const statusText = scene === 'donut' ? '🍩 Torus Scene Active' : '🌌 Asteroid Field Active';
            document.getElementById('statusText').textContent = statusText;
            
            // Send scene change to server
            fetch('/api/scene', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({scene: scene})
            });
        }

        function updateTorusParam(param, value) {
            // Update display
            document.getElementById(param + 'Value').textContent = value;
            
            // Send to server
            fetch('/api/params', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    [param]: parseFloat(value) || parseInt(value)
                })
            });
        }

        function updateAsteroidParam(param, value) {
            // Update display
            document.getElementById(param + 'Value').textContent = value;
            
            // Send to server
            fetch('/api/asteroid-params', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    [param]: parseFloat(value) || parseInt(value)
                })
            });
        }

        function resetTorusDefaults() {
            const defaults = {
                majorRadius: 2.0,
                minorRadius: 0.5,
                majorSegments: 50,
                minorSegments: 30,
                rotationSpeedX: 0.3,
                rotationSpeedY: 0.5
            };
            
            for (const [param, value] of Object.entries(defaults)) {
                document.getElementById(param).value = value;
                document.getElementById(param + 'Value').textContent = value;
            }
            
            fetch('/api/params', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(defaults)
            });
        }

        function resetAsteroidDefaults() {
            const defaults = {
                spawnRate: 1.0,
                maxAsteroids: 500,
                minSize: 0.05,
                maxSize: 0.3,
                minSpeed: 0.5,
                maxSpeed: 3.0,
                rotationSpeed: 1.0
            };
            
            for (const [param, value] of Object.entries(defaults)) {
                document.getElementById(param).value = value;
                document.getElementById(param + 'Value').textContent = value;
            }
            
            fetch('/api/asteroid-params', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(defaults)
            });
        }

        function randomizeTorus() {
            const params = {
                majorRadius: (Math.random() * 4 + 1).toFixed(1),
                minorRadius: (Math.random() * 1.5 + 0.2).toFixed(2),
                majorSegments: Math.floor(Math.random() * 60 + 20),
                minorSegments: Math.floor(Math.random() * 40 + 10),
                rotationSpeedX: (Math.random() * 2 - 1).toFixed(1),
                rotationSpeedY: (Math.random() * 2 - 1).toFixed(1)
            };
            
            for (const [param, value] of Object.entries(params)) {
                document.getElementById(param).value = value;
                document.getElementById(param + 'Value').textContent = value;
            }
            
            fetch('/api/params', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(params)
            });
        }

        function clearAsteroids() {
            fetch('/api/clear-asteroids', {
                method: 'POST'
            });
        }

        // Initialize WebSocket connection
        connectWebSocket();

        // Poll for updates every second
        setInterval(() => {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('renderMode').textContent = 
                        data.cudaEnabled ? 'CUDA Accelerated' : 'CPU Fallback';
                })
                .catch(err => console.log('Status check failed:', err));
        }, 1000);
    </script>
</body>
</html>
`

// StartWebServer starts the web interface server
func StartWebServer(params *TorusParams, cudaEnabled *bool) {
	// Update global reference
	if params != nil {
		currentParams = *params
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		return
	}

	// WebSocket endpoint for logs
	http.HandleFunc("/ws/logs", wsLogHandler)

	// Serve the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			TorusParams
			AsteroidParams
			CudaEnabled bool
		}{
			TorusParams:    currentParams,
			AsteroidParams: currentAsteroidParams,
			CudaEnabled:    cudaEnabled != nil && *cudaEnabled,
		}
		tmpl.Execute(w, data)
	})

	// API endpoint for scene switching
	http.HandleFunc("/api/scene", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var sceneData map[string]string
		if err := json.NewDecoder(r.Body).Decode(&sceneData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if scene, ok := sceneData["scene"]; ok {
			currentScene = scene
			LogSuccess("Scene", fmt.Sprintf("Scene switched to: %s", scene))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "scene changed"})
	})

	// API endpoint to update torus parameters
	http.HandleFunc("/api/params", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log parameter changes
		for key, value := range updates {
			LogInfo("UI", fmt.Sprintf("Torus parameter changed: %s = %v", key, value))
		}

		// Update parameters
		for key, value := range updates {
			switch key {
			case "majorRadius":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.MajorRadius
					currentParams.MajorRadius = float32(v)
					LogSuccess("Parameters", fmt.Sprintf("Major Radius: %.2f → %.2f", oldValue, float32(v)))
				}
			case "minorRadius":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.MinorRadius
					currentParams.MinorRadius = float32(v)
					LogSuccess("Parameters", fmt.Sprintf("Minor Radius: %.3f → %.3f", oldValue, float32(v)))
				}
			case "majorSegments":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.MajorSegments
					currentParams.MajorSegments = int(v)
					LogSuccess("Parameters", fmt.Sprintf("Major Segments: %d → %d", oldValue, int(v)))
				}
			case "minorSegments":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.MinorSegments
					currentParams.MinorSegments = int(v)
					LogSuccess("Parameters", fmt.Sprintf("Minor Segments: %d → %d", oldValue, int(v)))
				}
			case "rotationSpeedX":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.RotationSpeedX
					currentParams.RotationSpeedX = float32(v)
					LogSuccess("Parameters", fmt.Sprintf("X Rotation Speed: %.1f → %.1f", oldValue, float32(v)))
				}
			case "rotationSpeedY":
				if v, ok := value.(float64); ok {
					oldValue := currentParams.RotationSpeedY
					currentParams.RotationSpeedY = float32(v)
					LogSuccess("Parameters", fmt.Sprintf("Y Rotation Speed: %.1f → %.1f", oldValue, float32(v)))
				}
			}
		}

		// Copy back to the main params if provided
		if params != nil {
			*params = currentParams
		}

		LogInfo("System", "Torus parameters updated successfully")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	})

	// API endpoint to update asteroid parameters
	http.HandleFunc("/api/asteroid-params", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log parameter changes
		for key, value := range updates {
			LogInfo("UI", fmt.Sprintf("Asteroid parameter changed: %s = %v", key, value))
		}

		// Update parameters
		for key, value := range updates {
			switch key {
			case "spawnRate":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.SpawnRate
					currentAsteroidParams.SpawnRate = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Spawn Rate: %.1f → %.1f", oldValue, float32(v)))
				}
			case "maxAsteroids":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.MaxAsteroids
					currentAsteroidParams.MaxAsteroids = int(v)
					LogSuccess("Asteroids", fmt.Sprintf("Max Asteroids: %d → %d", oldValue, int(v)))
				}
			case "minSize":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.MinSize
					currentAsteroidParams.MinSize = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Min Size: %.3f → %.3f", oldValue, float32(v)))
				}
			case "maxSize":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.MaxSize
					currentAsteroidParams.MaxSize = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Max Size: %.3f → %.3f", oldValue, float32(v)))
				}
			case "minSpeed":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.MinSpeed
					currentAsteroidParams.MinSpeed = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Min Speed: %.1f → %.1f", oldValue, float32(v)))
				}
			case "maxSpeed":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.MaxSpeed
					currentAsteroidParams.MaxSpeed = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Max Speed: %.1f → %.1f", oldValue, float32(v)))
				}
			case "rotationSpeed":
				if v, ok := value.(float64); ok {
					oldValue := currentAsteroidParams.RotationSpeed
					currentAsteroidParams.RotationSpeed = float32(v)
					LogSuccess("Asteroids", fmt.Sprintf("Rotation Speed: %.1f → %.1f", oldValue, float32(v)))
				}
			}
		}

		LogInfo("System", "Asteroid parameters updated successfully")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	})

	// API endpoint to clear asteroids
	http.HandleFunc("/api/clear-asteroids", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		LogInfo("Asteroids", "Clearing all asteroids from field")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "asteroids cleared"})
	})

	// API endpoint for status
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cudaEnabled":    cudaEnabled != nil && *cudaEnabled,
			"params":         currentParams,
			"asteroidParams": currentAsteroidParams,
			"currentScene":   currentScene,
		})
	})

	LogInfo("WebServer", "Starting web interface on http://localhost:8080")
	LogInfo("WebServer", "Access the control panel and logs in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GetCurrentParams returns the current torus parameters
func GetCurrentParams() TorusParams {
	return currentParams
}

// GetCurrentAsteroidParams returns the current asteroid parameters
func GetCurrentAsteroidParams() AsteroidParams {
	return currentAsteroidParams
}

// GetCurrentScene returns the current scene name
func GetCurrentScene() string {
	return currentScene
}
