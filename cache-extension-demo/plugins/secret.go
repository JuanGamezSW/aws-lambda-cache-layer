package plugins

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Struct for storing secrets cache
type SecretConfiguration struct {
	Region string
	Names  []string
}

// Struct for caching the information
type Secret struct {
	CacheData CacheData
	Region    string
}

var secretCache = make(map[string]Secret)
var secretClient = make(map[string]*secretsmanager.SecretsManager)

// Initialize map and cache objects (if requested)
func InitSecret(secrets []SecretConfiguration, initializeCache bool) {
	for _, config := range secrets {
		for _, secret := range config.Names {
			_, isSecretPresent := secretCache[secret]
			if !isSecretPresent {
				if initializeCache {
					// Read from SecretManager and add it to the cache
					GetSecret(secret, GetRegion(config.Region), GetSecretManagerClient(GetRegion(config.Region)))
				} else {
					secretCache[secret] = Secret{
						CacheData: CacheData{},
						Region:    GetRegion(config.Region),
					}
				}
			} else {
				println(PrintPrefix, secret+" already exists so skipping it")
			}
		}
	}
}

// Initialize SecretManager cache
func GetSecret(name string, region string, smClient *secretsmanager.SecretsManager) string {
	result, err := smClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				println(PrintPrefix, secretsmanager.ErrCodeDecryptionFailure, err.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				println(PrintPrefix, secretsmanager.ErrCodeInternalServiceError, err.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				println(PrintPrefix, secretsmanager.ErrCodeInvalidParameterException, err.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				println(PrintPrefix, secretsmanager.ErrCodeInvalidRequestException, err.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				println(PrintPrefix, secretsmanager.ErrCodeResourceNotFoundException, err.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			println(PrintPrefix, "Error while fetching secret ", name, err.Error())
		}
		return ""
	} else {
		// Read Secret value from SecretManager
		var value string
		if result.SecretString != nil {
			value = *result.SecretString
		} else {
			decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
			len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
			if err != nil {
				println(PrintPrefix, "Base64 Decode Error ", err.Error())
				return ""
			}
			value = string(decodedBinarySecretBytes[:len])
		}
		// Update cache with value
		secretCache[name] = Secret{
			CacheData: CacheData{
				Data:        value,
				CacheExpiry: GetCacheExpiry(),
			},
			Region: region,
		}
		return value
	}
}

// Get SecretManager Client and cache it based on region
func GetSecretManagerClient(region string) *secretsmanager.SecretsManager {
	secretManagerClient, isCachePresent := secretClient[region]
	if !isCachePresent {
		sess, err := session.NewSessionWithOptions(session.Options{
			Config:            aws.Config{Region: aws.String(region)},
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			panic(err)
		}
		secretManagerClient = secretsmanager.New(sess, aws.NewConfig().WithRegion(region))
		secretClient[region] = secretManagerClient
	}

	return secretManagerClient
}

// Fetch Secret cache
func GetSecretCache(name string) string {
	var secret = secretCache[name]

	// If expired or not available in cache then read it from SecretManager, else return from cache
	if secret.CacheData.Data == "" || IsExpired(secret.CacheData.CacheExpiry) {
		println(PrintPrefix, "Fetching secret from AWS.")
		return GetSecret(name, GetRegion(secret.Region), GetSecretManagerClient(GetRegion(secret.Region)))
	} else {
		return secretCache[name].CacheData.Data
	}
}
