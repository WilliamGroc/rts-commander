#!/bin/bash

# Script d'installation pour Raspberry Pi

echo "🔧 Installation de RTS Commander"
echo ""

# Vérifier qu'on est sur une Raspberry Pi
if ! grep -q "Raspberry Pi" /proc/cpuinfo 2>/dev/null; then
    echo "⚠️  Attention: Ce script est conçu pour Raspberry Pi"
    read -p "Continuer quand même ? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Activer le SPI si nécessaire
if ! lsmod | grep -q spi_bcm2835; then
    echo "📡 Activation du SPI..."
    sudo raspi-config nonint do_spi 0
    echo "✅ SPI activé (un redémarrage peut être nécessaire)"
fi

# Installer Go si nécessaire
if ! command -v go &> /dev/null; then
    echo "📦 Installation de Go..."
    wget https://go.dev/dl/go1.24.3.linux-armv6l.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.24.3.linux-armv6l.tar.gz
    rm go1.24.3.linux-armv6l.tar.gz
    
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    echo "✅ Go installé"
fi

# Compiler l'application
echo "🔨 Compilation de RTS Commander..."
go build -o rtsCommander main.go

if [ $? -eq 0 ]; then
    echo "✅ Compilation réussie"
    
    # Créer un fichier de config exemple
    if [ ! -f remotes.json ]; then
        echo "📝 Création du fichier de configuration..."
        cp remotes.example.json remotes.json
        echo "✅ Fichier remotes.json créé"
    fi
    
    # Créer un service systemd
    echo "🔧 Création du service systemd..."
    sudo tee /etc/systemd/system/rtscommander.service > /dev/null <<EOF
[Unit]
Description=RTS Commander - Somfy Shutters Controller
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/rtsCommander --http :8080
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    echo "✅ Service créé"
    
    echo ""
    echo "🎉 Installation terminée !"
    echo ""
    echo "📋 Prochaines étapes :"
    echo "1. Éditer remotes.json pour ajouter vos volets"
    echo "   ou utiliser: ./rtsCommander --add --remote <nom> --address <adresse>"
    echo ""
    echo "2. Démarrer le service:"
    echo "   sudo systemctl start rtscommander"
    echo "   sudo systemctl enable rtscommander  # Pour démarrage automatique"
    echo ""
    echo "3. Vérifier le statut:"
    echo "   sudo systemctl status rtscommander"
    echo ""
    echo "4. API disponible sur: http://$(hostname -I | awk '{print $1}'):8080"
    echo ""
    echo "💡 Mode manuel (sans service):"
    echo "   ./rtsCommander --remote <nom> --cmd <up|down|my|prog>"
    echo ""
    
else
    echo "❌ Erreur lors de la compilation"
    exit 1
fi
