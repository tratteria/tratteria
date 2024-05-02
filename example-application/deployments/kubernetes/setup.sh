echo "Enable SPIRE? (y/n)"
read enable_spire

if [[ "$enable_spire" =~ ^[Yy]$ ]]; then
    ENABLE_SPIRE="true"
elif [[ "$enable_spire" =~ ^[Nn]$ ]]; then
    ENABLE_SPIRE="false"
else
    echo "Invalid input. Please enter 'y' or 'n'."
    exit 1
fi

echo "Enable TxnToken? (y/n)"
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

ENABLE_ACCESS_EVALUATION="false"

if [[ "$ENABLE_TXN_TOKEN" == "true" ]]; then
    echo "Enable Access Evaluation API in Tratteria? (y/n)"
    read enable_access_evaluation_api

    if [[ "$enable_access_evaluation_api" =~ ^[Yy]$ ]]; then
        ENABLE_ACCESS_EVALUATION="true"

        read -sp "Enter the Access Evaluation API bearer token: " ACCESS_EVALUATION_API_BEARER_TOKEN
        echo
        export ACCESS_EVALUATION_API_BEARER_TOKEN
    elif [[ "$enable_access_evaluation_api" =~ ^[Nn]$ ]]; then
        ENABLE_ACCESS_EVALUATION="false"
    else
        echo "Invalid input. Please enter 'y' or 'n'."
        exit 1
    fi
fi

export ENABLE_ACCESS_EVALUATION