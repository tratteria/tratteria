# Building Tnx-Token images
echo "\nBuilding Tnx-Token Service Image...\n"

docker build -t txn-token:latest -f ../../../../service/Dockerfile ../../../../service

# Applying the Txn-Token Configurations
echo "\nApplying the Txn-Token Configurations...\n"

# Generating Transaction Tokens Signing Keys
echo "\nGenerating Transaction Tokens Signing Keys...\n"

PRIVATE_KEY=$(openssl genrsa 2048)

PUBLIC_KEY=$(echo "$PRIVATE_KEY" | openssl rsa -pubout)

ENCODED_PRIVATE_KEY=$(echo "$PRIVATE_KEY" | base64 | tr -d '\n')

KEY_ID=$(echo "$PUBLIC_KEY" | openssl rsa -pubin -outform der | openssl dgst -sha256 -binary | base64 | tr -d '\n')

MODULUS=$(echo "$PUBLIC_KEY" | openssl rsa -pubin -modulus -noout | cut -d'=' -f2 | xxd -r -p | base64 | tr '/+' '_-' | tr -d '=')
EXPONENT="AQAB"

JWKS=$(cat <<EOF
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "$KEY_ID",
      "alg": "RS256",
      "n": "$MODULUS",
      "e": "$EXPONENT"
    }
  ]
}
EOF
)

kubectl create secret -n txn-token-ns generic rsa-keys \
  --from-literal=privateKey="$ENCODED_PRIVATE_KEY" \
  --from-literal=jwks="$JWKS" \
  --from-literal=KeyID="$KEY_ID"

kubectl create configmap txn-token-service-public-key --from-literal=jwks="$JWKS" -n alpha-stocks-dev

kubectl create configmap config --from-file=config.yaml=configs/config.yaml -n txn-token-ns
  
kubectl apply -f roles/
kubectl apply -f service-accounts/
kubectl apply -f deployments/
kubectl apply -f services
