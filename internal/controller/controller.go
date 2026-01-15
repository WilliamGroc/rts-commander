package controller

import (
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/conn/v3/spi"

	"rtscommander/m/internal/config"
	"rtscommander/m/internal/radio"
	"rtscommander/m/internal/remote"
)

// Controller gère l'envoi de commandes RTS
type Controller struct {
	config *config.Config
	conn   spi.Conn
	mu     sync.Mutex
}

// New crée un nouveau contrôleur
func New(cfg *config.Config, conn spi.Conn) *Controller {
	return &Controller{
		config: cfg,
		conn:   conn,
	}
}

// SendCommand envoie une commande RTS complète
func (ctrl *Controller) SendCommand(remoteName string, command byte) error {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()

	// Récupérer la télécommande
	rc, exists := ctrl.config.GetRemote(remoteName)

	if !exists {
		return fmt.Errorf("remote '%s' not found", remoteName)
	}

	// Créer la trame
	frame := rc.BuildRTSFrame(command)

	// Préambule et sync word Somfy
	preamble := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA} // Préambule
	syncWord := []byte{0xA0}                               // Hardware sync

	// Encoder en Manchester
	encodedFrame := remote.ManchesterEncode(frame)

	// Construire la trame complète
	fullFrame := append(preamble, syncWord...)
	fullFrame = append(fullFrame, encodedFrame...)

	// Mettre le CC1101 en mode idle
	if err := radio.WriteStrobe(ctrl.conn, radio.SIDLE); err != nil {
		return fmt.Errorf("failed to set idle mode: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// Vider le FIFO TX
	if err := radio.WriteStrobe(ctrl.conn, radio.SFTX); err != nil {
		return fmt.Errorf("failed to flush TX FIFO: %v", err)
	}

	// Envoyer la trame (répétition standard Somfy : 2 trames complètes + 7 répétitions)
	for i := 0; i < 2; i++ {
		if err := radio.TransmitFrame(ctrl.conn, fullFrame); err != nil {
			return fmt.Errorf("failed to send frame %d: %v", i+1, err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	// Répétitions avec inter-frame spacing
	for i := 0; i < 7; i++ {
		if err := radio.TransmitFrame(ctrl.conn, fullFrame); err != nil {
			return fmt.Errorf("failed to send repeat %d: %v", i+1, err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	// Incrémenter le rolling code
	rc.RollingCode++

	// Sauvegarder la configuration
	if err := ctrl.config.Save(); err != nil {
		log.Printf("Warning: failed to save config: %v", err)
	}

	log.Printf("[%s] Commande 0x%X envoyée (rolling code: %d)", remoteName, command, rc.RollingCode-1)
	return nil
}

// Config retourne la configuration du contrôleur
func (ctrl *Controller) Config() *config.Config {
	return ctrl.config
}
