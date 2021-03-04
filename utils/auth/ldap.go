package auth

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"golang.org/x/crypto/bcrypt"
	"jefembe.co.tz/vas/smsgw/utils/logger"
	"jefembe.co.tz/vas/smsgw/utils/propertymanager"
)

func AuthUsingUPN(username, password string) error {
	ldapEndpoint := propertymanager.GetStringProperty("ldap.jefembe.endpoint")
	conn, err := ldap.Dial("tcp", ldapEndpoint)
	if err != nil {
		logger.Error("Failed to connect: %v, %s", err, username)
		return err
	}
	//userPrincipalName[jefembe.co]
	upn := username + propertymanager.GetStringProperty("ldap.jefembe.upn-suffix")
	if err := conn.Bind(upn, password); err != nil {
		logger.Error("Authentication failed: %v, %s", err, username)
		return err
	}
	return nil
}

func AuthUsingDN(username, password string) error {
	adminDN := propertymanager.GetStringProperty("ldap.search.admin-dn")
	adminPassword := propertymanager.GetStringProperty("ldap.search.admin-password")
	userDn := ""
	ldapEndpoint := propertymanager.GetStringProperty("ldap.jefembe.endpoint")
	conn, err := ldap.Dial("tcp", ldapEndpoint)
	if err != nil {
		logger.Error("Failed to connect: %v, %s", err, username)
		return err
	}
	if err := conn.Bind(adminDN, adminPassword); err != nil {
		logger.Error("Admin bind failed with: %v, %s", err, adminDN)
		return err
	}
	searchRequest := ldap.NewSearchRequest(
		propertymanager.GetStringProperty("ldap.search.base-dn"),
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		propertymanager.GetIntProperty("ldap.search.timeout", 0),
		false,
		fmt.Sprintf("(uid=%s)", username),
		[]string{"distinguishedName"},
		nil,
	)
	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		logger.Error("User search failed: %v, %s", err, username)
		return err
	}
	if len(searchResult.Entries) != 1 {
		logger.Error("User does not exists or too many entries: %s", username)
		return errors.New("User does not exists or too many entries")
	}
	userDn = searchResult.Entries[0].GetAttributeValue("distinguishedName") //OR searchResult.Entries[0].DN
	logger.Info("Got DN as: %s", userDn)
	if err := conn.Bind(userDn, password); err != nil {
		logger.Error("Authentication failed: %v, %s", err, userDn)
		return err
	}
	logger.Info("Successfully authenticated against LDAP: %s", username)
	return nil
}

func AuthUsingLocal(username, password string) error {
	user, err := UserExists(username)
	if err != nil {
		return errors.New("Failed to authenticate")
	}
	if err == nil && user != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			logger.Error("Failure: %v", err)
			return err
		}
		logger.Info("Password matched: %s", username)
		return nil
	}
	logger.Info("User %s is not authorised", username)
	return errors.New("User not authorised")
}

func AuthenticateUser(username, password string) error {
	user, err := UserExists(username)
	if err != nil {
		return err
	}
	if err == nil && user != nil {
		useLdap := propertymanager.GetBoolProperty("ldap.auth.enabled")
		if useLdap {
			useUpn := propertymanager.GetBoolProperty("ldap.jefembe.use-upn")
			if useUpn {
				return AuthUsingUPN(username, password)
			}
			return AuthUsingDN(username, password)
		} else {
			return AuthUsingLocal(username, password)
		}
	}
	return errors.New("User not authorised")
}
