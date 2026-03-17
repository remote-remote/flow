package config

import "github.com/zalando/go-keyring"

const serviceName = "flow"

func SetSecret(key, value string) error {
	return keyring.Set(serviceName, key, value)
}

func GetSecret(key string) (string, error) {
	return keyring.Get(serviceName, key)
}
