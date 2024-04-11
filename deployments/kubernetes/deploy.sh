# Deploying Spire
echo "\nDeploying Spire...\n"

cd spire
./deploy.sh
cd ..

# Deploying Alpha Stocks
echo "\nDeploying Alpha Stocks...\n"

cd alpha-stocks
./deploy.sh
cd ..