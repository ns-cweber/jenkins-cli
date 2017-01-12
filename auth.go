package main

import (
	"fmt"
	"os/user"

	"github.com/howeyc/gopass"
	"github.com/zalando/go-keyring"
)

type credentials struct{ username, password string }

func getCredentials(prompt string) (credentials, error) {
	const service string = "ns_ldap"

	user, err := user.Current()
	if err != nil {
		return credentials{}, err
	}

	password, err := keyring.Get(service, user.Username)
	if err != nil {
		if err == keyring.ErrNotFound {
			fmt.Print(prompt)
			data, err := gopass.GetPasswd()
			if err != nil {
				return credentials{}, err
			}

			password := string(data)
			err = keyring.Set(service, user.Username, password)
			return credentials{user.Username, password}, err
		}
		return credentials{}, err
	}
	return credentials{user.Username, password}, nil
}
