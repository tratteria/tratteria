# Destroying Alpha Stock Resources
echo "\nDestroying Alpha Stock Resources...\n"

cd alpha-stocks
chmod +x destroy.sh
./destroy.sh
cd ..

# Destroying Spire Resources
echo "\n\n\nDestroying Spire Resources...\n"

cd spire
chmod +x destroy.sh
./destroy.sh
cd ..

# Destroying Tratteria Resources
echo "\nDestroying Tratteria Resources...\n"

cd tratteria
chmod +x destroy.sh
./destroy.sh
cd ..

# Destroying Dex Resources
echo "\nDestroying Dex Resources...\n"

cd dex
chmod +x destroy.sh
./destroy.sh
cd ..