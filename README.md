# RTS Commander - Contrôleur de volets roulants Somfy

Application Go pour contrôler des volets roulants Somfy RTS via une Raspberry Pi et un module CC1101.

## 🔧 Matériel requis

- Raspberry Pi (3/4/Zero)
- Module CC1101 (433.42 MHz)
- Connexion SPI activée sur la Raspberry Pi

## 📦 Installation

```bash
# Cloner le projet
git clone <votre-repo>
cd rtsCommander

# Compiler
go build -o rtsCommander ./cmd/rtscommander

# Ou compiler pour Raspberry Pi depuis un autre système
GOOS=linux GOARCH=arm GOARM=7 go build -o rtsCommander ./cmd/rtscommander
```

## 🚀 Utilisation

### 1. Ajouter une télécommande virtuelle

```bash
# Ajouter un volet nommé "salon" avec une adresse unique
./rtsCommander --add --remote salon --address 0x123456

# Avec paramètres personnalisés
./rtsCommander --add --remote chambre --address 0x223344 --rolling 1 --key 0xA7
```

⚠️ **Important**: Chaque télécommande virtuelle doit avoir une adresse unique (24 bits, entre 0x000001 et 0xFFFFFF).

### 2. Appairer la télécommande virtuelle

Une fois créée, vous devez appairer la télécommande avec votre volet :

```bash
# Envoyer la commande de programmation
./rtsCommander --remote salon --cmd prog

# Puis immédiatement appuyer sur le bouton PROG de votre télécommande physique
# pendant 3 secondes. Le volet fera un va-et-vient pour confirmer.
```

### 3. Envoyer des commandes

```bash
# Monter le volet
./rtsCommander --remote salon --cmd up

# Descendre le volet
./rtsCommander --remote salon --cmd down

# Stop / Position favorite
./rtsCommander --remote salon --cmd my
```

### 4. Lister les télécommandes configurées

```bash
./rtsCommander --list
```

### 5. Mode serveur HTTP (API REST)

```bash
# Démarrer le serveur sur le port 8080
./rtsCommander --http :8080
```

## 🌐 API HTTP

Une fois le serveur démarré, vous pouvez contrôler vos volets via HTTP :

### Envoyer une commande

```bash
curl -X POST http://localhost:8080/command \
  -H "Content-Type: application/json" \
  -d '{"remote": "salon", "command": "up"}'
```

Commandes disponibles : `up`, `down`, `my`, `stop`, `prog`

### Lister les télécommandes

```bash
curl http://localhost:8080/remotes
```

### Obtenir les détails d'une télécommande

```bash
curl http://localhost:8080/remote?name=salon
```

### Ajouter une télécommande via l'API

```bash
curl -X POST http://localhost:8080/remote/add \
  -H "Content-Type: application/json" \
  -d '{
    "name": "terrasse",
    "address": 4456789,
    "rolling_code": 1,
    "encryption_key": 167
  }'
```

## 📁 Fichier de configuration

Le fichier `remotes.json` stocke vos télécommandes virtuelles et leur rolling code :

```json
{
  "salon": {
    "name": "salon",
    "address": 1193046,
    "rolling_code": 45,
    "encryption_key": 167
  },
  "chambre": {
    "name": "chambre",
    "address": 2236723,
    "rolling_code": 12,
    "encryption_key": 167
  }
}
```

⚠️ **Ne perdez pas ce fichier !** Le rolling code doit être incrémenté à chaque commande pour des raisons de sécurité.

## 🏠 Intégration Home Assistant

Exemple de configuration avec Home Assistant :

```yaml
# configuration.yaml
cover:
  - platform: command_line
    covers:
      salon:
        command_open: 'curl -X POST http://raspberrypi:8080/command -H ''Content-Type: application/json'' -d ''{"remote":"salon","command":"up"}'''
        command_close: 'curl -X POST http://raspberrypi:8080/command -H ''Content-Type: application/json'' -d ''{"remote":"salon","command":"down"}'''
        command_stop: 'curl -X POST http://raspberrypi:8080/command -H ''Content-Type: application/json'' -d ''{"remote":"salon","command":"my"}'''
```

## 🐳 Docker

Créez un `Dockerfile` pour faciliter le déploiement :

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o rtsCommander main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rtsCommander .
VOLUME /root/config
CMD ["./rtsCommander", "--http", ":8080", "--config", "/root/config/remotes.json"]
```

## 🔒 Sécurité

- Chaque télécommande virtuelle a une adresse unique
- Le rolling code empêche la réplication des commandes
- Le fichier de configuration doit être protégé (contient les adresses et rolling codes)

## 🐛 Dépannage

### Le volet ne répond pas

1. Vérifiez que le module CC1101 est correctement connecté
2. Vérifiez que le SPI est activé : `sudo raspi-config` → Interface Options → SPI
3. Assurez-vous que la télécommande virtuelle est bien appairée
4. Vérifiez les logs pour d'éventuelles erreurs

### Rolling code désynchronisé

Si le volet ne répond plus après avoir perdu le fichier `remotes.json` :

1. Recréez la télécommande avec une nouvelle adresse
2. Appairez-la à nouveau avec `--cmd prog`

### Tester le module CC1101

```bash
# Vérifier que le module est détecté
ls /dev/spidev*
# Devrait afficher : /dev/spidev0.0 et/ou /dev/spidev0.1
```

## 📚 Références

- [Protocole Somfy RTS](https://pushstack.wordpress.com/somfy-rts-protocol/)
- [Module CC1101](https://www.ti.com/product/CC1101)
- [Périph.io Documentation](https://periph.io/)

## 📝 Licence

MIT License

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir une issue ou une pull request.
