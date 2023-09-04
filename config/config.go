// config/config.go

package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"os"
)

var (
	RpcAddr       string
	ChainId       string
	PrivateKey    string
	PrivateAESKey []byte
	KeyValutURL   string
)

func Init() {
	KeyValutURL = "https://bahenfileservice.vault.azure.net"
	RpcAddr, _ = GetSecretFromVault(KeyValutURL, "RpcAddr")
	ChainId, _ = GetSecretFromVault(KeyValutURL, "ChainId")
	PrivateKey, _ = GetSecretFromVault(KeyValutURL, "PrivateKey")
	PrivateAESKeyStr, _ := GetSecretFromVault(KeyValutURL, "PrivateAESKey")
	PrivateAESKey, _ = base64.StdEncoding.DecodeString(PrivateAESKeyStr)
	//PrivateAESKey, _ = generateAESKey(256)
}

// GetSecretFromVault retrieves a secret string from Azure Key Vault or returns dummy values if running locally.
func GetSecretFromVault(vaultURL string, secretName string) (string, error) {
	// Check for an environment variable to determine if we're running locally
	if os.Getenv("RUNNING_Azure") == "" {
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
		return "RpcAddr"
	case "ChainId":
		return "ChainId"
	case "PrivateKey":
		return "PrivateKey"
	case "PrivateAESKey":
		return "PrivateAESKey"
	default:
		return "unknown-secret"
	}
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

func getEncryptionKeyFromEnv() ([]byte, error) {
	encodedKey := os.Getenv("AES_KEY_ENV_VAR")
	if encodedKey == "" {
		return nil, errors.New("encryption key not found in environment variable")
	}

	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}
