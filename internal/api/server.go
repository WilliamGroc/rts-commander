package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"rtscommander/m/internal/controller"
	"rtscommander/m/internal/remote"
)

// CommandRequest représente une requête de commande
type CommandRequest struct {
	Remote  string `json:"remote"`
	Command string `json:"command"`
}

// CommandResponse représente une réponse de commande
type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Remote  string `json:"remote,omitempty"`
}

// Server représente le serveur HTTP
type Server struct {
	ctrl *controller.Controller
}

// NewServer crée un nouveau serveur API
func NewServer(ctrl *controller.Controller) *Server {
	return &Server{ctrl: ctrl}
}

// handleCommand gère les requêtes d'envoi de commande
func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convertir la commande string en byte
	var cmdByte byte
	switch req.Command {
	case "up", "UP", "monter":
		cmdByte = remote.CmdUp
	case "down", "DOWN", "descendre":
		cmdByte = remote.CmdDown
	case "my", "MY", "stop", "STOP":
		cmdByte = remote.CmdMy
	case "prog", "PROG", "program":
		cmdByte = remote.CmdProg
	default:
		sendJSONError(w, fmt.Sprintf("Unknown command: %s", req.Command), http.StatusBadRequest)
		return
	}

	// Envoyer la commande
	if err := s.ctrl.SendCommand(req.Remote, cmdByte); err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, CommandResponse{
		Success: true,
		Message: fmt.Sprintf("Command '%s' sent to '%s'", req.Command, req.Remote),
		Remote:  req.Remote,
	})
}

// handleListRemotes liste toutes les télécommandes
func (s *Server) handleListRemotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	remotes := s.ctrl.Config().ListRemotes()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"remotes": remotes,
		"count":   len(remotes),
	})
}

// handleGetRemote obtient les détails d'une télécommande
func (s *Server) handleGetRemote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		sendJSONError(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	remote, exists := s.ctrl.Config().GetRemote(name)

	if !exists {
		sendJSONError(w, fmt.Sprintf("Remote '%s' not found", name), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(remote)
}

// handleAddRemote ajoute une nouvelle télécommande
func (s *Server) handleAddRemote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rc remote.Control
	if err := json.NewDecoder(r.Body).Decode(&rc); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if rc.Name == "" {
		sendJSONError(w, "Missing remote name", http.StatusBadRequest)
		return
	}

	if err := s.ctrl.Config().AddRemote(rc.Name, &rc); err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, CommandResponse{
		Success: true,
		Message: fmt.Sprintf("Remote '%s' added successfully", rc.Name),
		Remote:  rc.Name,
	})
}

// sendJSONResponse envoie une réponse JSON
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// sendJSONError envoie une erreur JSON
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(CommandResponse{
		Success: false,
		Message: message,
	})
}

// Start démarre le serveur HTTP
func (s *Server) Start(addr string) error {
	http.HandleFunc("/command", s.handleCommand)
	http.HandleFunc("/remotes", s.handleListRemotes)
	http.HandleFunc("/remote", s.handleGetRemote)
	http.HandleFunc("/remote/add", s.handleAddRemote)

	log.Printf("HTTP server starting on %s", addr)
	log.Println("Endpoints:")
	log.Println("  POST   /command       - Send a command")
	log.Println("  GET    /remotes       - List all remotes")
	log.Println("  GET    /remote?name=X - Get remote details")
	log.Println("  POST   /remote/add    - Add a new remote")

	return http.ListenAndServe(addr, nil)
}
