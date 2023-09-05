// config/config.go

package config

import (
	"crypto/rand"
	"os"
)

var (
	RpcAddr       string = "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	ChainId       string = "greenfield_5600-1"
	PrivateKey    string = os.Getenv("PRIVATE_KEY")
	PrivateAESKey []byte = []byte("L2aOTHL2SfuKSX1GygHMd2/QdboGYkpfPSGmpROTRiJGDKVqYUkh8BlHZdsT1JM1")
	KeyValutURL   string = "https://decenterfileservice.vault.azure.net"
)

// FIXME: getEncryptionKeyfromEnv

func getEnvVar(holder *string, envVar string) {
	e := os.Getenv(envVar)
	if e != "" {
		*holder = e
	}
}

func init() {
	getEnvVar(&ChainId, "ChainId")
	getEnvVar(&RpcAddr, "RpcAddr")
	getEnvVar(&PrivateKey, "PrivateKey")

	// RpcAddr, _ = GetSecretFromVault(KeyValutURL, "RpcAddr")
	// ChainId, _ = GetSecretFromVault(KeyValutURL, "ChainId")
	// PrivateKey, _ = GetSecretFromVault(KeyValutURL, "PrivateKey")
	// PrivateAESKeyStr, _ := GetSecretFromVault(KeyValutURL, "PrivateAESKey")
	// PrivateAESKey, _ = base64.StdEncoding.DecodeString(PrivateAESKeyStr)
	// PrivateAESKey, _ = generateAESKey(256)
}

func generateAESKey(bits int) ([]byte, error) {
	keyLength := bits / 8 // 8 bits in a byte
	key := make([]byte, keyLength)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

/*func getEncryptionKeyFromEnv() ([]byte, error) {
	encodedKey := os.Getenv("AES_KEY_ENV_VAR")
	if encodedKey == "" {
		return nil, errors.New("encryption key not found in environment variable")
	}

	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}*/

/*// GetSecretFromVault retrieves a secret string from Azure Key Vault or returns dummy values if running locally.
func GetSecretFromVault(vaultURL string, secretName string) (string, error) {
	// Check for an environment variable to determine if we're running locally
	if os.Getenv("mode") == "" || os.Getenv("mode") == "local" {
		return getLocalDummyValue(secretName), nil
	}

	client := keyvault.New()

	// Use the Azure environment based authentication
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return "", err
	}
	client.Authorizer = authorizer

	// Construct the full secret URL
	secretURL := vaultURL + "/secrets/" + secretName

	secretBundle, err := client.GetSecret(context.Background(), KeyValutURL, secretURL, "")
	if err != nil {
		return "", err
	}

	return *secretBundle.Value, nil
}

// getLocalDummyValue returns a dummy value for the given secret name when running locally.
func getLocalDummyValue(secretName string) string {
	// replace these value to true value.
	switch secretName {
	case "RpcAddr":
		return "https://gnfd-testnet-sp1.bnbchain.org:443"
	case "ChainId":
		return "bnbchain-sp1"
	case "PrivateKey":
		return "" // FIXME: private key
	case "PrivateAESKey":
		return "L2aOTHL2SfuKSX1GygHMd2/QdboGYkpfPSGmpROTRiJGDKVqYUkh8BlHZdsT1JM1"
	default:
		return "unknown-secret"
	}
}*/
