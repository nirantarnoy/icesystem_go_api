package config

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

const (
	MapUser = "administrator@cicnetgrp.net"
	MapPass = "Tamagogi@&2019$"
	FQDN    = "172.16.0.205"
	BaseDN  = "cn=Configuration,dc=cicnetgrp,dc=net"
	Filter  = "(objectClass=*)"
)

func Connect() (*ldap.Conn, error) {
	ldap, err := ldap.DialURL(fmt.Sprintf("ldap://%s:389", FQDN))
	if err != nil {
		return nil, err
	}
	return ldap, nil
}
