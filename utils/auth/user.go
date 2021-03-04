package auth

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"time"

	"jefembe.co.tz/vas/smsgw/utils/logger"
	"jefembe.co.tz/vas/smsgw/utils/propertymanager"
)

type Users struct {
	XMLName xml.Name `xml:"users"`
	User    []User   `xml:"user"`
}

//var usersMap map[string]User

type User struct {
	XMLName      xml.Name `xml:"user"`
	FullName     string   `xml:"fullName"`
	Username     string   `xml:"username"`
	Password     string   `xml:"password"`
	Admin        bool     `xml:"admin"`
	GroupId      string   `xml:"groupId"`
	CreationDate string   `xml:"creationDate"`
	LastLogin    string   `xml:"lastLogin"`
}

func UserExists(username string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users.User {
		if user.Username == username {
			logger.Info("User %s found with groupId: %s", username, user.GroupId)
			return &user, nil
		}
	}
	logger.Warn("User %s not found", username)
	return nil, errors.New("User not found")
}

func GetAllUsers() (*Users, error) {
	usersFile, err := os.Open(propertymanager.GetStringProperty("users.file-location", "./conf/users.xml"))
	defer usersFile.Close()
	if err != nil {
		logger.Error("Failed to open file: %v", err)
		return nil, err
	}
	usersInBytes, err := ioutil.ReadAll(usersFile)
	if err != nil {
		logger.Error("Failed to read users file: %v", err)
		return nil, err
	}
	var users Users
	if err := xml.Unmarshal(usersInBytes, &users); err != nil {
		logger.Error("Failed to Unmarshal: %v", err)
		return nil, err
	}
	return &users, nil
}

func UserAdd(user User) error {
	_, err := UserExists(user.Username)
	if err == nil {
		return errors.New("User already exists")
	}
	users, err := GetAllUsers()
	if err != nil {
		logger.Error("Failed to get users: %v", err)
		return err
	}

	users.User = append(users.User, user)

	file, err := os.Create(propertymanager.GetStringProperty("users.file-location", "./conf/users.xml"))
	if err != nil {
		logger.Error("Failed to create file: %v", err)
		return err
	}
	xmlWriter := io.Writer(file)
	encoder := xml.NewEncoder(xmlWriter)
	encoder.Indent(" ", "  ")
	if err := encoder.Encode(users); err != nil {
		logger.Error("Failed to encode XML user data into a file: %v", err)
		return err
	}
	return nil
}

func UserRemove(username string) error {
	_, err := UserExists(username)
	if err != nil {
		return err
	}
	users, err := GetAllUsers()
	if err != nil {
		logger.Error("Failed to get users: %v", err)
		return err
	}

	usersMap := make(map[string]User)
	for _, v := range users.User {
		usersMap[v.Username] = v
	}
	delete(usersMap, username)
	updatedUsers := &Users{}
	for _, v := range usersMap {
		updatedUsers.User = append(updatedUsers.User, v)
	}

	usersFile, err := os.Create(propertymanager.GetStringProperty("users.file-location", "./conf/users.xml"))
	if err != nil {
		logger.Error("Failed to create users file: %v", err)
		return err
	}
	usersXmlWriter := io.Writer(usersFile)
	usersXmlEncoder := xml.NewEncoder(usersXmlWriter)
	usersXmlEncoder.Indent(" ", "  ")
	if err := usersXmlEncoder.Encode(updatedUsers); err != nil {
		logger.Error("Failed to encode XML user data into a file: %v", err)
		return err
	}
	defer usersFile.Close()
	return nil
}

func UserUpdateLastLogin(username string) error {
	users, err := GetAllUsers()
	if err != nil {
		logger.Error("Failed to get users: %v", err)
		return err
	}

	usersMap := make(map[string]User)
	for _, v := range users.User {
		usersMap[v.Username] = v
	}

	usersMap[username] = User{
		FullName:     usersMap[username].FullName,
		Username:     usersMap[username].Username,
		Password:     usersMap[username].Password,
		Admin:        usersMap[username].Admin,
		GroupId:      usersMap[username].GroupId,
		CreationDate: usersMap[username].CreationDate,
		LastLogin:    time.Now().Format("2006-01-02 15:04:05"),
	}
	updatedUsers := &Users{}
	for _, v := range usersMap {
		updatedUsers.User = append(updatedUsers.User, v)
	}
	usersFile, err := os.Create(propertymanager.GetStringProperty("users.file-location", "./conf/users.xml"))
	if err != nil {
		logger.Error("Failed to create users file: %v", err)
		return err
	}
	usersXmlWriter := io.Writer(usersFile)
	usersXmlEncoder := xml.NewEncoder(usersXmlWriter)
	usersXmlEncoder.Indent(" ", "  ")
	if err := usersXmlEncoder.Encode(updatedUsers); err != nil {
		logger.Error("Failed to encode XML user data into a file: %v", err)
		return err
	}
	return nil
}

// func main() {
// 	fmt.Println(UserUpdateLastLogin("natayg"))
// }
