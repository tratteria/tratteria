echo "Toggle on SPIRE? (y/n)"
read enable_spire

if [[ "$enable_spire" =~ ^[Yy]$ ]]; then
    ENABLE_SPIRE="true"
elif [[ "$enable_spire" =~ ^[Nn]$ ]]; then
    ENABLE_SPIRE="false"
else
    echo "Invalid input. Please enter 'y' or 'n'."
    exit 1
fi

echo "Toggle on TxnToken ? (y/n)"
read enable_txn_token

if [[ "$enable_txn_token" =~ ^[Yy]$ ]]; then
    ENABLE_TXN_TOKEN="true"
elif [[ "$enable_txn_token" =~ ^[Nn]$ ]]; then
    ENABLE_TXN_TOKEN="false"
else
    echo "Invalid input. Please enter 'y' or 'n'."
    exit 1
fi

export ENABLE_SPIRE
export ENABLE_TXN_TOKEN

echo "\n\n\nRedeploying Alpha Stocks...\n\n\n"

cd alpha-stocks
chmod +x deploy.sh
./deploy.sh
cd ..
