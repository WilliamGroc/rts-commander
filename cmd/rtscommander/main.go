package main

import (
	"flag"
	"fmt"
	"log"

	"rtscommander/m/internal/api"
	"rtscommander/m/internal/config"
	"rtscommander/m/internal/controller"
	"rtscommander/m/internal/radio"
	"rtscommander/m/internal/remote"
)

func main() {
	// Flags CLI
	configPath := flag.String("config", "remotes.json", "Path to the configuration file")
	httpAddr := flag.String("http", "", "HTTP server address (e.g., :8080)")
	remoteName := flag.String("remote", "", "Remote control name")
	command := flag.String("cmd", "", "Command to send: up, down, my/stop, prog")
	addRemote := flag.Bool("add", false, "Add a new remote")
	listRemotes := flag.Bool("list", false, "List all remotes")
	address := flag.Uint("address", 0, "Remote address (24-bit, required for -add)")
	rollingCode := flag.Uint("rolling", 1, "Initial rolling code (for -add)")
	encKey := flag.Uint("key", 0xA7, "Encryption key (for -add)")

	flag.Parse()

	// Charger la configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Mode liste
	if *listRemotes {
		remotes := cfg.ListRemotes()
		if len(remotes) == 0 {
			fmt.Println("No remotes configured")
		} else {
			fmt.Printf("Configured remotes (%d):\n", len(remotes))
			for _, name := range remotes {
				r := cfg.Remotes[name]
				fmt.Printf("  - %s: address=0x%06X, rolling_code=%d\n",
					name, r.Address, r.RollingCode)
			}
		}
		return
	}

	// Mode ajout de télécommande
	if *addRemote {
		if *remoteName == "" || *address == 0 {
			log.Fatal("Usage: --add --remote <name> --address <addr>")
		}

		rc := &remote.Control{
			Name:          *remoteName,
			Address:       uint32(*address),
			RollingCode:   uint16(*rollingCode),
			EncryptionKey: byte(*encKey),
		}

		if err := cfg.AddRemote(*remoteName, rc); err != nil {
			log.Fatalf("Failed to add remote: %v", err)
		}

		fmt.Printf("Remote '%s' added successfully\n", *remoteName)
		fmt.Printf("  Address: 0x%06X\n", rc.Address)
		fmt.Printf("  Rolling Code: %d\n", rc.RollingCode)
		fmt.Printf("  Encryption Key: 0x%02X\n", rc.EncryptionKey)
		return
	}

	// Initialiser le CC1101
	conn, err := radio.InitCC1101()
	if err != nil {
		log.Fatalf("Failed to initialize CC1101: %v", err)
	}
	// Note: spi.Conn n'a pas de méthode Close, la connexion est gérée par periph.io

	// Créer le contrôleur
	ctrl := controller.New(cfg, conn)

	// Mode serveur HTTP
	if *httpAddr != "" {
		server := api.NewServer(ctrl)
		log.Fatal(server.Start(*httpAddr))
	}

	// Mode CLI - envoi de commande unique
	if *remoteName != "" && *command != "" {
		var cmdByte byte
		switch *command {
		case "up", "monter":
			cmdByte = remote.CmdUp
		case "down", "descendre":
			cmdByte = remote.CmdDown
		case "my", "stop":
			cmdByte = remote.CmdMy
		case "prog", "program":
			cmdByte = remote.CmdProg
		default:
			log.Fatalf("Unknown command: %s (use: up, down, my, stop, prog)", *command)
		}

		if err := ctrl.SendCommand(*remoteName, cmdByte); err != nil {
			log.Fatalf("Failed to send command: %v", err)
		}

		fmt.Printf("Command '%s' sent to '%s' successfully!\n", *command, *remoteName)
		return
	}

	// Aucune action spécifiée
	fmt.Println("Usage:")
	fmt.Println("  List remotes:    --list")
	fmt.Println("  Add remote:      --add --remote <name> --address <addr>")
	fmt.Println("  Send command:    --remote <name> --cmd <up|down|my|prog>")
	fmt.Println("  Start HTTP API:  --http :8080")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  Add a remote:")
	fmt.Println("    ./rtsCommander --add --remote salon --address 0x123456")
	fmt.Println("  Send a command:")
	fmt.Println("    ./rtsCommander --remote salon --cmd up")
	fmt.Println("  Start HTTP server:")
	fmt.Println("    ./rtsCommander --http :8080")
}
