package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"jefembe.co.tz/vas/smsgw/utils"
	"jefembe.co.tz/vas/smsgw/utils/auth"
	"jefembe.co.tz/vas/smsgw/utils/logger"
	"jefembe.co.tz/vas/smsgw/utils/propertymanager"
	"jefembe.co.tz/vas/smsgw/utils/smsc"
)

type CsvParams struct {
	Message  string
	SenderId string
	Line     []string
	Username string
}

type User struct {
	UserName string
}

func main() {
	http.HandleFunc("/smsgw", Smsgw)
	http.HandleFunc("/smsgwHandler", SmsgwHandler)
	http.HandleFunc("/uploadCSV", UploadCSV)
	http.HandleFunc("/getSenders", GetSenderAddresses)
	http.HandleFunc("/getUsers", GetUsers)
	http.HandleFunc("/userAdd", UserAdd)
	http.HandleFunc("/userRemove", UserRemove)
	http.HandleFunc("/getMessageSubmitted", GetMessagesSubmitted)

	serverPort := net.JoinHostPort("", propertymanager.GetStringProperty("server.http.port", "8090"))
	log.Printf("Prepare server to listen on %s", serverPort)
	http.ListenAndServe(serverPort, nil)
}
func UploadCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fileHandle, _, err := r.FormFile("fileName")
		message := r.FormValue("message")
		senderId := r.FormValue("senderId")
		username := r.FormValue("name")
		if err != nil {
			logger.Error("Failed to upload file: %v", err)
			http.Error(w, "Error in Uploading a file: ", http.StatusInternalServerError)
			return
		}
		defer fileHandle.Close()
		lines, err := csv.NewReader(fileHandle).ReadAll()
		if err != nil {
			logger.Error("Failed to read a file: %v", err)
			http.Error(w, "Error in Reading a file: ", http.StatusInternalServerError)
			return
		}
		msgArray := strings.Split(message, "$")
		if len(msgArray) != len(lines[0]) {
			fmt.Fprintln(w, "Number of columns must match the number of splitted message parts", http.StatusInternalServerError)
			logger.Error("Number of columns must match the number of splitted message parts")
			return
		}
		for _, line := range lines {
			csvParams := &CsvParams{
				Message:  message,
				SenderId: senderId,
				Line:     line,
				Username: username,
			}
			CsvProcessor(csvParams)
		}
		fmt.Fprintln(w, "Submittted by: ", username)
	}
}

func CsvProcessor(csvParams *CsvParams) {
	msgArray := strings.Split(csvParams.Message, "$")
	var messageCombo string = ""
	var msisdn = csvParams.Line[0]
	for i := range msgArray {
		if i+1 <= len(csvParams.Line)-1 {
			messageCombo += msgArray[i] + csvParams.Line[i+1]
		} else {
			messageCombo += msgArray[len(msgArray)-1]
		}
	}
	logger.Info("Concatenated combo message: %s", messageCombo)
	go smsc.SendSms(msisdn, messageCombo, csvParams.SenderId)
}

func SmsgwHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	logger.Info("Authenticate user %s", username)
	err := auth.AuthenticateUser(username, password)
	if err == nil {
		userDetails, _ := auth.UserExists(username)
		user := User{UserName: username}
		if userDetails.Admin {
			t := template.Must(template.ParseFiles("static/admin_user.html"))
			t.Execute(w, user)
		} else {
			t := template.Must(template.ParseFiles("static/normal_user.html"))
			t.Execute(w, user)
		}
		//Update lastLogin
		auth.UserUpdateLastLogin(username)
		logger.Info("Success login: %s", username)
		return
	}
	fmt.Fprintln(w, err)
}

func Smsgw(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/smsgw.html"))
	t.Execute(w, nil)
}

func GetSenderAddresses(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	senders := propertymanager.GetStringProperty("smsc.default.senderid")
	user, _ := auth.UserExists(username)
	if user.GroupId != "" {
		sendersArray := utils.GetSenders(user.GroupId)
		if sendersArray != nil {
			senders = strings.Join(sendersArray, "|")
		}
	}
	sendersJson := `{"type":"senders","senders":"` + senders + `"}`
	fmt.Fprintln(w, sendersJson)
	logger.Info("Return sender addresses: %s", sendersJson)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("static/user_mgt.html"))
	users, _ := auth.GetAllUsers()
	t.Execute(w, *users)
}

func UserAdd(w http.ResponseWriter, r *http.Request) {
	addUsername := r.FormValue("add_username")
	addPassword := r.FormValue("add_password")
	addGroupId := r.FormValue("add_groupId")
	addFullName := r.FormValue("add_fullName")
	addAdmin, _ := strconv.ParseBool(r.FormValue("add_admin"))
	addedBy := r.FormValue("name")

	if addFullName == "" || addUsername == "" || addPassword == "" || addGroupId == "" || r.FormValue("add_admin") == "" {
		fmt.Fprintln(w, "All fields are required!")
		return
	}

	hashPassword, err := utils.EncryptPassword(addPassword)
	if err != nil {
		fmt.Fprintln(w, "Failed to encrypt password, please retry")
		return
	}

	user := auth.User{
		FullName:     addFullName,
		Username:     addUsername,
		Password:     hashPassword,
		GroupId:      addGroupId,
		Admin:        addAdmin,
		CreationDate: time.Now().Format("2006-01-02 15:04:05"),
		LastLogin:    "",
	}

	if err := auth.UserAdd(user); err != nil {
		logger.Error("Failed to add user: %v, %v", err, user)
		fmt.Fprintln(w, err)
		return
	}
	logger.Info("User %s added by %s successfully", addUsername, addedBy)
	fmt.Fprintln(w, "User added successfully")
}

func UserRemove(w http.ResponseWriter, r *http.Request) {
	removedBy := r.URL.Query().Get("removedBy")
	username := r.URL.Query().Get("username")
	if err := auth.UserRemove(username); err != nil {
		fmt.Fprintf(w, "Failed to Delete user: %s\n%v", username, err)
		return
	}
	logger.Info("User %s deleted by %s successfully", username, removedBy)
	fmt.Fprintf(w, "User %s deleted successfully", username)
}

func GetMessagesSubmitted(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"data":[]}`)
}
