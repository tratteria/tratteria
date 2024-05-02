# Building Tratteria images
echo "\nBuilding Tratteria Image...\n"

docker build -t tratteria:latest -f ../../../../service/Dockerfile ../../../../service

# Applying the Tratteria Configurations
echo "\nApplying the Tratteria Service Configurations...\n"

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

kubectl create secret -n tts-ns generic rsa-keys \
  --from-literal=privateKey="$ENCODED_PRIVATE_KEY" \
  --from-literal=jwks="$JWKS" \
  --from-literal=KeyID="$KEY_ID"

kubectl create secret -n tts-ns generic access-evaluation-api-authentication --from-literal=ACCESS_EVALUATION_API_BEARER_TOKEN="$ACCESS_EVALUATION_API_BEARER_TOKEN"

kubectl create configmap tts-public-key --from-literal=jwks="$JWKS" -n alpha-stocks

kubectl create configmap config --from-file=config.yaml=configs/config.yaml -n tts-ns
  
kubectl apply -f roles/
kubectl apply -f service-accounts/
envsubst < ./deployments/txn-token-deployment.yaml | kubectl apply -f -
kubectl apply -f services
