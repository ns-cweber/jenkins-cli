package auth

import (
	"fmt"
	"os/user"

	"github.com/howeyc/gopass"
	"github.com/zalando/go-keyring"
)

type Credentials struct{ Username, Password string }

func GetCredentials(prompt string) (Credentials, error) {
	const service string = "ns_ldap"

	user, err := user.Current()
	if err != nil {
		return Credentials{}, err
	}

	password, err := keyring.Get(service, user.Username)
	if err != nil {
		if err == keyring.ErrNotFound {
			fmt.Print(prompt)
			data, err := gopass.GetPasswd()
			if err != nil {
				return Credentials{}, err
			}

			password := string(data)
			err = keyring.Set(service, user.Username, password)
			return Credentials{user.Username, password}, err
		}
		return Credentials{}, err
	}
	return Credentials{user.Username, password}, nil
}
