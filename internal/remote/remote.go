package remote

import (
	"encoding/binary"
)

// Commandes RTS Somfy
const (
	CmdMy   = 0x1 // Stop / My position
	CmdUp   = 0x2 // Monter
	CmdDown = 0x4 // Descendre
	CmdProg = 0x8 // Programmation
)

// Control représente une télécommande virtuelle
type Control struct {
	Name          string `json:"name"`
	Address       uint32 `json:"address"`        // Adresse de la télécommande (24 bits)
	RollingCode   uint16 `json:"rolling_code"`   // Rolling code (16 bits)
	EncryptionKey byte   `json:"encryption_key"` // Clé d'obfuscation
}

// BuildRTSFrame crée une trame RTS Somfy
func (rc *Control) BuildRTSFrame(command byte) []byte {
	// Trame Somfy RTS : 56 bits + sync
	frame := make([]byte, 7)

	// Octet 0 : clé (8 bits)
	frame[0] = rc.EncryptionKey

	// Octet 1 : commande (4 bits) + checksum (4 bits)
	ctrl := command << 4
	checksum := (ctrl ^ (ctrl >> 4) ^ byte(rc.RollingCode) ^ byte(rc.RollingCode>>8))
	frame[1] = ctrl | (checksum & 0x0F)

	// Octets 2-3 : rolling code (16 bits)
	binary.BigEndian.PutUint16(frame[2:4], rc.RollingCode)

	// Octets 4-6 : adresse (24 bits)
	frame[4] = byte(rc.Address >> 16)
	frame[5] = byte(rc.Address >> 8)
	frame[6] = byte(rc.Address)

	// Obfuscation de la trame
	for i := 1; i < 7; i++ {
		frame[i] ^= frame[i-1]
	}

	return frame
}

// ManchesterEncode encode une trame en Manchester
func ManchesterEncode(data []byte) []byte {
	encoded := make([]byte, 0, len(data)*2)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if (b>>uint(i))&1 == 1 {
				encoded = append(encoded, 0xAA) // 1 -> 10
			} else {
				encoded = append(encoded, 0x55) // 0 -> 01
			}
		}
	}
	return encoded
}
