FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copier les fichiers du projet
COPY go.mod go.sum* ./
RUN go mod download

COPY main.go ./

# Compiler l'application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rtsCommander main.go

# Image finale
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copier l'exécutable
COPY --from=builder /app/rtsCommander .

# Créer le répertoire de configuration
RUN mkdir -p /root/config

# Volume pour la configuration
VOLUME ["/root/config"]

# Port HTTP
EXPOSE 8080

# Lancer l'application
CMD ["./rtsCommander", "--http", ":8080", "--config", "/root/config/remotes.json"]
