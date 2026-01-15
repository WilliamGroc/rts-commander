# Structure du projet rtsCommander

## 📁 Organisation des fichiers

Le projet a été découpé en plusieurs fichiers pour améliorer la lisibilité et la maintenabilité :

### Fichiers principaux

- **main.go** (3.6 KB)

  - Point d'entrée de l'application
  - Gestion des arguments CLI (flags)
  - Logique principale du programme

- **config.go** (1.8 KB)

  - Structure `Config`
  - Chargement et sauvegarde de la configuration JSON
  - Gestion des télécommandes virtuelles

- **remote.go** (1.7 KB)

  - Structure `RemoteControl`
  - Constantes des commandes RTS (UP, DOWN, MY, PROG)
  - Construction et encodage des trames RTS Somfy
  - Encodage Manchester

- **cc1101.go** (3.8 KB)

  - Configuration du module CC1101
  - Initialisation SPI
  - Fonctions de communication bas niveau (registres, strobes)
  - Transmission des trames radio

- **controller.go** (2.3 KB)

  - Structure `Controller`
  - Orchestration de l'envoi des commandes RTS
  - Gestion du rolling code
  - Répétitions des trames Somfy

- **api.go** (4.6 KB)
  - Serveur HTTP REST
  - Handlers pour les endpoints
  - Sérialisation JSON
  - Gestion des requêtes/réponses

## 🔗 Dépendances entre fichiers

```
main.go
  ├── config.go (LoadConfig)
  ├── cc1101.go (initCC1101)
  └── controller.go (NewController)
      ├── remote.go (buildRTSFrame, manchesterEncode)
      ├── cc1101.go (writeStrobe, transmitFrame)
      └── config.go (Save)

api.go
  └── controller.go (sendRTSCommand)
```

## 📊 Avantages de la modularité

1. **Lisibilité** : Chaque fichier a une responsabilité claire
2. **Maintenabilité** : Plus facile de trouver et modifier du code
3. **Testabilité** : Possibilité de tester chaque module indépendamment
4. **Réutilisabilité** : Les modules peuvent être utilisés séparément
5. **Collaboration** : Plusieurs développeurs peuvent travailler simultanément

## 🧪 Compilation

```bash
# Compilation normale
go build

# Tous les fichiers .go du package main sont automatiquement inclus
# Pas besoin de les spécifier individuellement
```

## 📝 Notes

- Tous les fichiers appartiennent au package `main`
- Les fonctions/structures en majuscule sont exportées (publiques)
- Les fonctions/structures en minuscule sont privées au package
- Le découpage respecte le principe de responsabilité unique (SRP)
