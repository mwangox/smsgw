package utils

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/bcrypt"
	"jefembe.co.tz/vas/smsgw/utils/logger"
)

type Addresses struct {
	XMLName      xml.Name `xml:"addresses"`
	AddressGroup []struct {
		ID     string   `xml:"id,attr"`
		Sender []string `xml:"sender"`
	} `xml:"address-group"`
}

func GetSenders(groupID string) []string {
	sendersFile, err := os.Open("conf/senders_id.xml")
	if err != nil {
		logger.Error("Failed to open file: %v", err)
		return nil
	}
	sendersInBytes, err := ioutil.ReadAll(sendersFile)
	if err != nil {
		logger.Error("Failed to read senders file: %v", err)
		return nil
	}

	var addr Addresses
	if err := xml.Unmarshal(sendersInBytes, &addr); err != nil {
		logger.Error("Failed to Unmarshal: %v", err)
		return nil
	}
	for _, v := range addr.AddressGroup {
		if v.ID == groupID {
			return v.Sender
		}
	}
	defer sendersFile.Close()
	return nil
}

func EncryptPassword(plainPassword string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(hashPassword), nil
}
