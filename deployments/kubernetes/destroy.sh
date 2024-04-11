# Destroying Alpha Stock Resources
echo "\nDestroying Alpha Stock Resources...\n"

cd alpha-stocks
./destroy.sh
cd ..

# Destroying Spire Resources
echo "\n\n\nDestroying Spire Resources...\n"

cd spire
./destroy.sh
cd ..
