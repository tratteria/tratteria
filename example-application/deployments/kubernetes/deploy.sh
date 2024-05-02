chmod +x setup.sh
source setup.sh

echo "\n\n\nDeploying Alpha Stocks...\n\n\n"

# Creating namespace
echo "\nCreating Namespace...\n"

kubectl apply -f namespaces.yaml

# Deploying Dex
echo "\nDeploying Dex...\n"

cd dex
chmod +x deploy.sh
./deploy.sh
cd ..

# Deploying Spire
echo "\nDeploying Spire...\n"

cd spire
chmod +x deploy.sh
./deploy.sh
cd ..

# Deploying Alpha Stocks
echo "\nDeploying Alpha Stocks...\n"

cd alpha-stocks
chmod +x deploy.sh
./deploy.sh
cd ..

# Deploying Tratteria
echo "\nDeploying Tratteria...\n"

cd tratteria
chmod +x deploy.sh
./deploy.sh
cd ..