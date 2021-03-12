package smsc

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"github.com/google/uuid"
	"jefembe.co.tz/vas/smsgw/utils/logger"
	"jefembe.co.tz/vas/smsgw/utils/propertymanager"
)

var tx *smpp.Transmitter
var conn smpp.ConnStatus
var messageQueue chan MessageDetails
var MaxQueueSizeReached = errors.New("Max messageQueue size reached, consider adding more accounts")

type MessageDetails struct {
	Msisdn   string
	Msg      string
	SenderId string
}

type Smsc struct {
	XMLName   xml.Name     `xml:"smsc"`
	SmscGroup []*SmscGroup `xml:"smscGroup"`
}

type SmscGroup struct {
	ID             string `xml:"id,attr"`
	SmppUsername   string `xml:"smppUsername"`
	SmppPassword   string `xml:"smppPassword"`
	SmppHost       string `xml:"smppHost"`
	SmppSystemType string `xml:"smppSystemType"`
}

func init() {
	messageQueue = make(chan MessageDetails)
	smppAccounts, err := LoadSmppAccounts()
	if err != nil {
		logger.Fatal("Failed to load smsc accounts: %v", err)
	}
	smppFailedCount := 0
	for k, v := range smppAccounts {
		//Bind to smsc
		tx, conn := GetTx(v)
		if conn.Status() != smpp.Connected {
			smppFailedCount++
			logger.Info("Skip sprawn goroutine [%d] with this account: [%v, %s, %d]", k, conn.Error(), v.SmppUsername, smppFailedCount)
			if smppFailedCount == len(smppAccounts) {
				log.Fatalf("Smpp client failed to start: [totalFailures=%d]", smppFailedCount)
			}
			continue
		}

		go func(goId int, v *SmscGroup) {

			logger.Info("Start goroutine [%d] with: [%s, %s, %v]", goId, tx.User, tx.Addr, v)
			for {
				for messageDetails := range messageQueue {
					logger.Info("Goroutine [%d] got message: %v", goId, messageDetails)
					SendSmsWithTx(messageDetails.Msisdn, messageDetails.Msg, messageDetails.SenderId, tx)
				}
			}
		}(k, v)
	}
}

func SendSms(msisdn, msg, senderId string) error {
	messageDetails := MessageDetails{
		Msisdn:   msisdn,
		Msg:      msg,
		SenderId: senderId,
	}
	// if len(messageQueue) == 200000 {
	// 	log.Printf("%v, %s", MaxQueueSizeReached, msisdn)
	// 	return MaxQueueSizeReached
	// }
	messageQueue <- messageDetails
	return nil
}

func SendSmsWithTx(msisdn, msg, senderId string, tx *smpp.Transmitter) {
	msisdns := []string{msisdn}
	guid := uuid.New().String()
	logger.Info("Send to smsc: [ uuid=%s, msisdn=%s, msg=%s ]", guid, msisdn, msg)
	sm, err := tx.Submit(&smpp.ShortMessage{
		Src:      senderId,
		DstList:  msisdns,
		Text:     pdutext.Raw(msg),
		Register: pdufield.NoDeliveryReceipt,
	})
	if err != nil {
		logger.Error("Failed to submit SMS: [ %v, %s, %s ]", err, msisdn, guid)
		return
	}
	logger.Info("Submitted successully: [ %s, %s, %s, %s ]", sm.RespID(), msisdn, guid, tx.User)
}

func GetTx(smscGroup *SmscGroup) (*smpp.Transmitter, smpp.ConnStatus) {
	tx := &smpp.Transmitter{
		Addr:       smscGroup.SmppHost,
		User:       smscGroup.SmppUsername,
		Passwd:     smscGroup.SmppPassword,
		SystemType: smscGroup.SmppSystemType,
	}
	// Create persistent connection, wait for the first status.
	conn := <-tx.Bind()
	logger.Info("Bind to smsc with: [ user=%s, status=%s ]", tx.User, conn.Status())
	return tx, conn
}

func LoadSmppAccounts() ([]*SmscGroup, error) {
	smscFile, err := os.Open(propertymanager.GetStringProperty("smsc.file-location", "./conf/smsc.xml"))
	if err != nil {
		logger.Error("Failed to open smsc file: %v", err)
		return nil, err
	}
	smscInBytes, err := ioutil.ReadAll(smscFile)
	if err != nil {
		logger.Error("Failed to read smsc file: %v", err)
		return nil, err
	}
	var smsc Smsc
	if err := xml.Unmarshal(smscInBytes, &smsc); err != nil {
		logger.Error("Failed to Unmarshal: %v", err)
		return nil, err
	}
	defer smscFile.Close()
	return smsc.SmscGroup, nil
}
