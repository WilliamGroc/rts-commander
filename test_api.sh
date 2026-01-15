#!/bin/bash

# Tests rapides de l'API

API_URL="http://localhost:8080"

echo "🧪 Tests de l'API RTS Commander"
echo ""

# Test 1: Lister les télécommandes
echo "📋 Test 1: Lister les télécommandes"
curl -s "$API_URL/remotes" | jq '.'
echo ""

# Test 2: Ajouter une télécommande de test
echo "➕ Test 2: Ajouter une télécommande de test"
curl -s -X POST "$API_URL/remote/add" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test",
    "address": 999999,
    "rolling_code": 1,
    "encryption_key": 167
  }' | jq '.'
echo ""

# Test 3: Récupérer les détails
echo "🔍 Test 3: Récupérer les détails de 'test'"
curl -s "$API_URL/remote?name=test" | jq '.'
echo ""

# Test 4: Envoyer une commande (simulation)
echo "📤 Test 4: Envoyer une commande UP"
curl -s -X POST "$API_URL/command" \
  -H "Content-Type: application/json" \
  -d '{
    "remote": "test",
    "command": "up"
  }' | jq '.'
echo ""

echo "✅ Tests terminés"
echo ""
echo "💡 Pour tester avec vos vrais volets:"
echo "   curl -X POST $API_URL/command -H 'Content-Type: application/json' -d '{\"remote\":\"salon\",\"command\":\"up\"}'"
