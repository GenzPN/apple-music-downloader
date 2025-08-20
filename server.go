package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"main/utils/ampapi"
	"main/utils/structs"
	"gopkg.in/yaml.v2"
)

// DownloadTask represents a download task
type DownloadTask struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Server represents the web server
type Server struct {
	tasks    map[string]*DownloadTask
	taskMux  sync.RWMutex
	port     string
	config   structs.ConfigSet
}

// NewServer creates a new server instance
func NewServer(port string) *Server {
	return &Server{
		tasks: make(map[string]*DownloadTask),
		port:  port,
	}
}

// generateTaskID generates a unique task ID
func (s *Server) generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// addTask adds a new download task
func (s *Server) addTask(url, downloadType, quality string) *DownloadTask {
	s.taskMux.Lock()
	defer s.taskMux.Unlock()

	task := &DownloadTask{
		ID:        s.generateTaskID(),
		URL:       url,
		Type:      downloadType,
		Status:    "pending",
		Progress:  0,
		Message:   "Task created",
		CreatedAt: time.Now(),
	}

	s.tasks[task.ID] = task
	return task
}

// updateTask updates task status
func (s *Server) updateTask(taskID, status string, progress int, message string) {
	s.taskMux.Lock()
	defer s.taskMux.Unlock()

	if task, exists := s.tasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.Message = message
		if status == "completed" || status == "failed" {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
}

// getTask returns a task by ID
func (s *Server) getTask(taskID string) (*DownloadTask, bool) {
	s.taskMux.RLock()
	defer s.taskMux.RUnlock()

	task, exists := s.tasks[taskID]
	return task, exists
}

// getAllTasks returns all tasks
func (s *Server) getAllTasks() []*DownloadTask {
	s.taskMux.RLock()
	defer s.taskMux.RUnlock()

	tasks := make([]*DownloadTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// handleHome serves the main page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.New("home").Parse(htmlTemplate))
	tmpl.Execute(w, nil)
}

// handleDownload handles download requests
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL    string `json:"url"`
		Type   string `json:"type"`
		Quality string `json:"quality"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate URL
	storefront, id := s.validateURL(req.URL)
	if storefront == "" || id == "" {
		http.Error(w, "Invalid Apple Music URL", http.StatusBadRequest)
		return
	}

	// Create task
	task := s.addTask(req.URL, req.Type, req.Quality)

	// Start download in background
	go s.processDownload(task, storefront, id, req.Quality)

	// Return task info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id": task.ID,
		"status":  "started",
	})
}

// validateURL validates and extracts storefront and ID from Apple Music URL
func (s *Server) validateURL(url string) (string, string) {
	// Album URL
	if strings.Contains(url, "/album/") {
		return checkUrl(url)
	}
	// Song URL
	if strings.Contains(url, "/song/") {
		return checkUrlSong(url)
	}
	// Playlist URL
	if strings.Contains(url, "/playlist/") {
		return checkUrlPlaylist(url)
	}
	// Artist URL
	if strings.Contains(url, "/artist/") {
		return checkUrlArtist(url)
	}
	// Music Video URL
	if strings.Contains(url, "/music-video/") {
		return checkUrlMv(url)
	}
	// Station URL
	if strings.Contains(url, "/station/") {
		return checkUrlStation(url)
	}
	return "", ""
}

// processDownload processes the download in background
func (s *Server) processDownload(task *DownloadTask, storefront, id, quality string) {
	s.updateTask(task.ID, "processing", 10, "Starting download...")

	// Get token
	token, err := ampapi.GetToken()
	if err != nil {
		if s.config.AuthorizationToken != "" && s.config.AuthorizationToken != "your-authorization-token" {
			token = strings.Replace(s.config.AuthorizationToken, "Bearer ", "", -1)
		} else {
			s.updateTask(task.ID, "failed", 0, "Failed to get authorization token")
			return
		}
	}

	s.updateTask(task.ID, "processing", 20, "Token obtained, analyzing content...")

	// Set quality flags
	s.setQualityFlags(quality)

	// Process based on type
	var err2 error
	if strings.Contains(task.URL, "/album/") {
		s.updateTask(task.ID, "processing", 30, "Downloading album...")
		err2 = ripAlbum(id, token, storefront, s.config.MediaUserToken, "")
	} else if strings.Contains(task.URL, "/song/") {
		s.updateTask(task.ID, "processing", 30, "Downloading song...")
		err2 = ripSong(id, token, storefront, s.config.MediaUserToken)
	} else if strings.Contains(task.URL, "/playlist/") {
		s.updateTask(task.ID, "processing", 30, "Downloading playlist...")
		err2 = ripPlaylist(id, token, storefront, s.config.MediaUserToken)
	} else if strings.Contains(task.URL, "/artist/") {
		s.updateTask(task.ID, "processing", 30, "Processing artist...")
		// For artist, we need to get albums first
		urlArtistName, urlArtistID, err := getUrlArtistName(task.URL, token)
		if err != nil {
			s.updateTask(task.ID, "failed", 0, "Failed to get artist information")
			return
		}
		
		// Set artist folder format
		Config.ArtistFolderFormat = strings.NewReplacer(
			"{UrlArtistName}", LimitString(urlArtistName),
			"{ArtistId}", urlArtistID,
		).Replace(Config.ArtistFolderFormat)
		
		// Get artist albums (simplified for web interface)
		albumArgs, err := checkArtist(task.URL, token, "albums")
		if err != nil {
			s.updateTask(task.ID, "failed", 0, "Failed to get artist albums")
			return
		}
		
		if len(albumArgs) == 0 {
			s.updateTask(task.ID, "failed", 0, "No albums found for this artist")
			return
		}
		
		s.updateTask(task.ID, "processing", 50, fmt.Sprintf("Downloading %d albums...", len(albumArgs)))
		
		// Download each album
		for i, albumURL := range albumArgs {
			progress := 50 + (i * 40 / len(albumArgs))
			s.updateTask(task.ID, "processing", progress, fmt.Sprintf("Downloading album %d of %d", i+1, len(albumArgs)))
			
			albumStorefront, albumID := checkUrl(albumURL)
			if albumStorefront != "" && albumID != "" {
				err = ripAlbum(albumID, token, albumStorefront, s.config.MediaUserToken, "")
				if err != nil {
					log.Printf("Failed to download album %s: %v", albumURL, err)
				}
			}
		}
	} else if strings.Contains(task.URL, "/music-video/") {
		s.updateTask(task.ID, "processing", 30, "Downloading music video...")
		if len(s.config.MediaUserToken) <= 50 {
			s.updateTask(task.ID, "failed", 0, "Media user token is required for music videos")
			return
		}
		err2 = mvDownloader(id, s.config.AlacSaveFolder, token, storefront, s.config.MediaUserToken, nil)
	} else if strings.Contains(task.URL, "/station/") {
		s.updateTask(task.ID, "processing", 30, "Downloading station...")
		if len(s.config.MediaUserToken) <= 50 {
			s.updateTask(task.ID, "failed", 0, "Media user token is required for stations")
			return
		}
		err2 = ripStation(id, token, storefront, s.config.MediaUserToken)
	} else {
		s.updateTask(task.ID, "failed", 0, "Unsupported URL type")
		return
	}

	if err2 != nil {
		s.updateTask(task.ID, "failed", 0, fmt.Sprintf("Download failed: %v", err2))
		return
	}

	s.updateTask(task.ID, "completed", 100, "Download completed successfully")
}

// setQualityFlags sets the global quality flags based on user selection
func (s *Server) setQualityFlags(quality string) {
	dl_atmos = false
	dl_aac = false

	switch quality {
	case "atmos":
		dl_atmos = true
	case "aac":
		dl_aac = true
		*aac_type = "aac"
	case "alac":
		// Default is ALAC
	}
}

// handleStatus returns task status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, exists := s.getTask(taskID)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// handleTasks returns all tasks
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.getAllTasks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// handleConfig returns current configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

// Start starts the web server
func (s *Server) Start() error {
	// Load configuration
	if err := s.loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Set up routes
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/api/download", s.handleDownload)
	http.HandleFunc("/api/status", s.handleStatus)
	http.HandleFunc("/api/tasks", s.handleTasks)
	http.HandleFunc("/api/config", s.handleConfig)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Starting Apple Music Downloader server on port %s", s.port)
	log.Printf("Open your browser and go to: http://localhost:%s", s.port)
	
	return http.ListenAndServe(":"+s.port, nil)
}

// loadConfig loads configuration from file
func (s *Server) loadConfig() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	
	// Use the existing Config variable from main.go
	if err := yaml.Unmarshal(data, &s.config); err != nil {
		return err
	}
	
	// Also set the global Config variable
	Config = s.config
	
	if len(s.config.Storefront) != 2 {
		s.config.Storefront = "us"
		Config.Storefront = "us"
	}
	
	return nil
}

// HTML template for the web interface
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Apple Music Downloader</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #ff6b6b, #ee5a24);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            font-weight: 700;
        }
        
        .header p {
            font-size: 1.1em;
            opacity: 0.9;
        }
        
        .content {
            padding: 40px;
        }
        
        .form-group {
            margin-bottom: 25px;
        }
        
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }
        
        input[type="url"], select {
            width: 100%;
            padding: 15px;
            border: 2px solid #e1e5e9;
            border-radius: 10px;
            font-size: 16px;
            transition: border-color 0.3s ease;
        }
        
        input[type="url"]:focus, select:focus {
            outline: none;
            border-color: #667eea;
        }
        
        .btn {
            background: linear-gradient(135deg, #667eea, #764ba2);
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s ease;
            width: 100%;
        }
        
        .btn:hover {
            transform: translateY(-2px);
        }
        
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        
        .tasks {
            margin-top: 40px;
        }
        
        .task {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 15px;
            border-left: 4px solid #667eea;
        }
        
        .task-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        
        .task-status {
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .status-pending { background: #ffeaa7; color: #d63031; }
        .status-processing { background: #74b9ff; color: white; }
        .status-completed { background: #00b894; color: white; }
        .status-failed { background: #e17055; color: white; }
        
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e1e5e9;
            border-radius: 4px;
            overflow: hidden;
            margin: 10px 0;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #667eea, #764ba2);
            transition: width 0.3s ease;
        }
        
        .task-message {
            color: #666;
            font-size: 14px;
        }
        
        .refresh-btn {
            background: #00b894;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 8px;
            cursor: pointer;
            margin-top: 20px;
        }
        
        .loading {
            display: none;
            text-align: center;
            margin: 20px 0;
        }
        
        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            width: 30px;
            height: 30px;
            animation: spin 1s linear infinite;
            margin: 0 auto 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🍎 Apple Music Downloader</h1>
            <p>Download your favorite music from Apple Music</p>
        </div>
        
        <div class="content">
            <form id="downloadForm">
                <div class="form-group">
                    <label for="url">Apple Music URL:</label>
                    <input type="url" id="url" name="url" placeholder="https://music.apple.com/..." required>
                </div>
                
                <div class="form-group">
                    <label for="quality">Audio Quality:</label>
                    <select id="quality" name="quality">
                        <option value="alac">Lossless (ALAC)</option>
                        <option value="aac">High-Quality (AAC)</option>
                        <option value="atmos">Dolby Atmos</option>
                    </select>
                </div>
                
                <button type="submit" class="btn" id="downloadBtn">
                    Start Download
                </button>
            </form>
            
            <div class="loading" id="loading">
                <div class="spinner"></div>
                <p>Processing download...</p>
            </div>
            
            <div class="tasks" id="tasks">
                <h3>Download Tasks</h3>
                <button class="refresh-btn" onclick="loadTasks()">Refresh Tasks</button>
                <div id="tasksList"></div>
            </div>
        </div>
    </div>

    <script>
        let currentTaskId = null;
        
        document.getElementById('downloadForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const url = document.getElementById('url').value;
            const quality = document.getElementById('quality').value;
            const downloadBtn = document.getElementById('downloadBtn');
            const loading = document.getElementById('loading');
            
            downloadBtn.disabled = true;
            loading.style.display = 'block';
            
            try {
                const response = await fetch('/api/download', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        url: url,
                        quality: quality
                    })
                });
                
                if (response.ok) {
                    const data = await response.json();
                    currentTaskId = data.task_id;
                    document.getElementById('url').value = '';
                    alert('Download started! Check the tasks below for progress.');
                    loadTasks();
                    startPolling(data.task_id);
                } else {
                    const error = await response.text();
                    alert('Error: ' + error);
                }
            } catch (error) {
                alert('Error: ' + error.message);
            } finally {
                downloadBtn.disabled = false;
                loading.style.display = 'none';
            }
        });
        
        function startPolling(taskId) {
            const interval = setInterval(async () => {
                try {
                    const response = await fetch('/api/status?task_id=' + taskId);
                    if (response.ok) {
                        const task = await response.json();
                        updateTaskDisplay(task);
                        
                        if (task.status === 'completed' || task.status === 'failed') {
                            clearInterval(interval);
                        }
                    }
                } catch (error) {
                    console.error('Error polling task status:', error);
                }
            }, 2000);
        }
        
        async function loadTasks() {
            try {
                const response = await fetch('/api/tasks');
                if (response.ok) {
                    const tasks = await response.json();
                    displayTasks(tasks);
                }
            } catch (error) {
                console.error('Error loading tasks:', error);
            }
        }
        
        function displayTasks(tasks) {
            const tasksList = document.getElementById('tasksList');
            tasksList.innerHTML = '';
            
            if (tasks.length === 0) {
                tasksList.innerHTML = '<p>No tasks yet.</p>';
                return;
            }
            
            tasks.forEach(task => {
                const taskElement = createTaskElement(task);
                tasksList.appendChild(taskElement);
            });
        }
        
        function createTaskElement(task) {
            const div = document.createElement('div');
            div.className = 'task';
            div.id = 'task-' + task.id;
            
            const statusClass = 'status-' + task.status;
            
            div.innerHTML = `
                <div class="task-header">
                    <strong>` + task.type + ` - ` + task.url.substring(0, 50) + `...</strong>
                    <span class="task-status ` + statusClass + `">` + task.status + `</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ` + task.progress + `%"></div>
                </div>
                <div class="task-message">` + task.message + `</div>
                <small>Created: ` + new Date(task.created_at).toLocaleString() + `</small>
            `;
            
            return div;
        }
        
        function updateTaskDisplay(task) {
            const taskElement = document.getElementById('task-' + task.id);
            if (taskElement) {
                const statusElement = taskElement.querySelector('.task-status');
                const progressElement = taskElement.querySelector('.progress-fill');
                const messageElement = taskElement.querySelector('.task-message');
                
                statusElement.className = 'task-status status-' + task.status;
                statusElement.textContent = task.status;
                progressElement.style.width = task.progress + '%';
                messageElement.textContent = task.message;
            }
        }
        
        // Load tasks on page load
        document.addEventListener('DOMContentLoaded', function() {
            loadTasks();
        });
    </script>
</body>
</html>
` 