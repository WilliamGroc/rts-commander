package radio

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

// Configuration SPI
const (
	spiSpeed = 1 * physic.MegaHertz
)

// Registres CC1101
const (
	WriteBurst = 0x40
	ReadSingle = 0x80
	ReadBurst  = 0xC0
	Strobe     = 0x30
	SRES       = 0x30 // Reset
	SIDLE      = 0x36 // Idle
	STX        = 0x35 // TX mode
	SFTX       = 0x3B // Flush TX FIFO
)

// Configuration du CC1101 pour Somfy RTS (433.42 MHz)
var cc1101Config = map[byte]byte{
	0x00: 0x0D, // IOCFG2
	0x01: 0x2E, // IOCFG1
	0x02: 0x2E, // IOCFG0
	0x03: 0x07, // FIFOTHR
	0x04: 0xD3, // SYNC1
	0x05: 0x91, // SYNC0
	0x06: 0xFF, // PKTLEN
	0x07: 0x04, // PKTCTRL1
	0x08: 0x32, // PKTCTRL0 - Mode asynchrone
	0x09: 0x00, // ADDR
	0x0A: 0x00, // CHANNR
	0x0B: 0x06, // FSCTRL1
	0x0C: 0x00, // FSCTRL0
	0x0D: 0x10, // FREQ2 - 433.42 MHz
	0x0E: 0xB0, // FREQ1
	0x0F: 0x71, // FREQ0
	0x10: 0xF8, // MDMCFG4 - Bande passante et débit
	0x11: 0x93, // MDMCFG3
	0x12: 0x03, // MDMCFG2 - Modulation ASK/OOK
	0x13: 0x22, // MDMCFG1
	0x14: 0xF8, // MDMCFG0
	0x15: 0x15, // DEVIATN
	0x18: 0x18, // MCSM0
	0x19: 0x16, // FOCCFG
	0x1B: 0x43, // AGCCTRL2
	0x1C: 0x40, // AGCCTRL1
	0x1D: 0x91, // AGCCTRL0
	0x21: 0x56, // FREND1
	0x22: 0x10, // FREND0
	0x23: 0xE9, // FSCAL3
	0x24: 0x2A, // FSCAL2
	0x25: 0x00, // FSCAL1
	0x26: 0x1F, // FSCAL0
	0x2C: 0x81, // TEST2
	0x2D: 0x35, // TEST1
	0x2E: 0x09, // TEST0
}

// InitCC1101 initialise le module CC1101
func InitCC1101() (spi.Conn, error) {
	// Ouvrir la connexion SPI
	port, err := spireg.Open("")
	if err != nil {
		return nil, fmt.Errorf("failed to open SPI: %v", err)
	}

	conn, err := port.Connect(spiSpeed, spi.Mode0, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SPI: %v", err)
	}

	// Reset du CC1101
	if err := WriteStrobe(conn, SRES); err != nil {
		return nil, fmt.Errorf("failed to reset CC1101: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Configuration des registres
	for addr, value := range cc1101Config {
		if err := WriteRegister(conn, addr, value); err != nil {
			return nil, fmt.Errorf("failed to configure register 0x%02X: %v", addr, err)
		}
	}

	// Vérifier que le CC1101 est bien configuré
	partnum, err := ReadRegister(conn, 0xF0) // PARTNUM
	if err != nil {
		return nil, fmt.Errorf("failed to read PARTNUM: %v", err)
	}
	if partnum != 0x00 {
		log.Printf("Warning: Unexpected PARTNUM value: 0x%02X (expected 0x00)", partnum)
	}

	log.Println("CC1101 initialisé avec succès")
	return conn, nil
}

// WriteRegister écrit dans un registre du CC1101
func WriteRegister(conn spi.Conn, addr, value byte) error {
	tx := []byte{addr, value}
	rx := make([]byte, len(tx))
	if err := conn.Tx(tx, rx); err != nil {
		return err
	}
	return nil
}

// TestCC1101 teste la communication avec le CC1101 et affiche les informations
func TestCC1101(conn spi.Conn) error {
	// Lire PARTNUM (devrait être 0x00 pour CC1101)
	partnum, err := ReadRegister(conn, 0xF0)
	if err != nil {
		return fmt.Errorf("failed to read PARTNUM register: %v", err)
	}
	fmt.Printf("  PARTNUM: 0x%02X ", partnum)
	if partnum == 0x00 {
		fmt.Println("✓ (CC1101 détecté)")
	} else {
		fmt.Printf("⚠ (valeur attendue: 0x00)\n")
	}

	// Lire VERSION
	version, err := ReadRegister(conn, 0xF1)
	if err != nil {
		return fmt.Errorf("failed to read VERSION register: %v", err)
	}
	fmt.Printf("  VERSION: 0x%02X ", version)
	if version == 0x04 || version == 0x14 {
		fmt.Println("✓")
	} else {
		fmt.Printf("⚠ (valeur typique: 0x04 ou 0x14)\n")
	}

	// Lire quelques registres de configuration pour vérifier la programmation
	fmt.Println("\n3. Vérification de la configuration...")
	
	// FREQ2, FREQ1, FREQ0 - Fréquence
	freq2, _ := ReadRegister(conn, 0x0D)
	freq1, _ := ReadRegister(conn, 0x0E)
	freq0, _ := ReadRegister(conn, 0x0F)
	fmt.Printf("  Fréquence (FREQ2/1/0): 0x%02X%02X%02X ", freq2, freq1, freq0)
	if freq2 == 0x10 && freq1 == 0xB0 && freq0 == 0x71 {
		fmt.Println("✓ (433.42 MHz)")
	} else {
		fmt.Println("⚠")
	}

	// MDMCFG2 - Configuration modulation
	mdmcfg2, _ := ReadRegister(conn, 0x12)
	fmt.Printf("  Modulation (MDMCFG2): 0x%02X ", mdmcfg2)
	if mdmcfg2 == 0x03 {
		fmt.Println("✓ (ASK/OOK)")
	} else {
		fmt.Println("⚠")
	}

	// PKTCTRL0 - Mode de paquet
	pktctrl0, _ := ReadRegister(conn, 0x08)
	fmt.Printf("  Mode paquet (PKTCTRL0): 0x%02X ", pktctrl0)
	if pktctrl0 == 0x32 {
		fmt.Println("✓ (Asynchrone)")
	} else {
		fmt.Println("⚠")
	}

	// Test de l'état (MARCSTATE)
	marcstate, err := ReadRegister(conn, 0xF5)
	if err != nil {
		return fmt.Errorf("failed to read MARCSTATE register: %v", err)
	}
	fmt.Printf("\n  État du module (MARCSTATE): 0x%02X ", marcstate)
	switch marcstate & 0x1F {
	case 0x01:
		fmt.Println("✓ (IDLE)")
	case 0x13:
		fmt.Println("✓ (TX)")
	case 0x14:
		fmt.Println("✓ (RX)")
	default:
		fmt.Printf("(État: %d)\n", marcstate&0x1F)
	}

	return nil
}

// ReadRegister lit un registre du CC1101
func ReadRegister(conn spi.Conn, addr byte) (byte, error) {
	tx := []byte{addr | ReadSingle, 0x00}
	rx := make([]byte, len(tx))
	if err := conn.Tx(tx, rx); err != nil {
		return 0, err
	}
	return rx[1], nil
}

// WriteStrobe envoie une commande strobe au CC1101
func WriteStrobe(conn spi.Conn, strobe byte) error {
	tx := []byte{strobe}
	rx := make([]byte, 1)
	return conn.Tx(tx, rx)
}

// TransmitFrame transmet une trame via le CC1101
func TransmitFrame(conn spi.Conn, frame []byte) error {
	// Écrire la trame dans le FIFO TX (burst write)
	tx := make([]byte, len(frame)+1)
	tx[0] = 0x3F | WriteBurst // FIFO address
	copy(tx[1:], frame)
	rx := make([]byte, len(tx))

	if err := conn.Tx(tx, rx); err != nil {
		return err
	}

	// Lancer la transmission
	if err := WriteStrobe(conn, STX); err != nil {
		return err
	}

	// Attendre la fin de la transmission
	time.Sleep(100 * time.Millisecond)

	return nil
}
