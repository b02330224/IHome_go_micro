// submail project submail.go
package submail

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

//base
func HttpGet(queryurl string) string {
	u, _ := url.Parse(queryurl)
	retstr, err := http.Get(u.String())
	if err != nil {
		return err.Error()
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}

func HttpPost(queryurl string, postdata map[string]string) string {
	data, err := json.Marshal(postdata)
	if err != nil {
		return err.Error()
	}

	body := bytes.NewBuffer([]byte(data))

	retstr, err := http.Post(queryurl, "application/json;charset=utf-8", body)

	if err != nil {
		return err.Error()
	}
	result, err := ioutil.ReadAll(retstr.Body)
	retstr.Body.Close()
	if err != nil {
		return err.Error()
	}
	return string(result)
}

func GetTimeStamp() string {
	resp := HttpGet("https://api.submail.cn/service/timestamp.json")
	var dict map[string]interface{}
	err := json.Unmarshal([]byte(resp), &dict)
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(int(dict["timestamp"].(float64)))
}

func CreateSignatrue(request map[string]string, config map[string]string) string {
	appkey := config["appkey"]
	appid := config["appid"]
	signtype := config["signtype"]
	request["sign_type"] = signtype
	keys := make([]string, 0, 32)
	for key, _ := range request {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	str_list := make([]string, 0, 32)
	for _, key := range keys {
		str_list = append(str_list, fmt.Sprintf("%s=%s", key, request[key]))
	}
	sigstr := strings.Join(str_list, "&")
	sigstr = fmt.Sprintf("%s%s%s%s%s", appid, appkey, sigstr, appid, appkey)
	if signtype == "normal" {
		return appkey
	} else if signtype == "md5" {
		mymd5 := md5.New()
		io.WriteString(mymd5, sigstr)
		return fmt.Sprintf("%x", mymd5.Sum(nil))
	} else {
		mysha1 := sha1.New()
		io.WriteString(mysha1, sigstr)
		return fmt.Sprintf("%x", mysha1.Sum(nil))
	}
}
func MailSendRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/mail/send.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MailXSendRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/mail/xsend.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MailSubscribeRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/addressbook/mail/subscribe.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MailUnSubscribeRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/addressbook/mail/unsubscribe.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MessageSendRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/message/send.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MessageXSendRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/message/xsend.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MessageSubscribeRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/addressbook/message/subscribe.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

func MessageUnSubscribeRun(request map[string]string, config map[string]string) string {
	url := "https://api.submail.cn/addressbook/message/unsubscribe.json"
	request["appid"] = config["appid"]
	request["timestamp"] = GetTimeStamp()
	request["signature"] = CreateSignatrue(request, config)
	return HttpPost(url, request)
}

//mailsend
type MailSend struct {
	to          []map[string]string
	addressbook []string
	from        string
	fromname    string
	reply       string
	cc          []map[string]string
	bcc         []map[string]string
	subject     string
	text        string
	html        string
	vars        map[string]string
	links       map[string]string
	headers     map[string]string
}

func CreateMailSend() *MailSend {
	mailsend := new(MailSend)
	mailsend.vars = make(map[string]string)
	mailsend.links = make(map[string]string)
	mailsend.headers = make(map[string]string)
	return mailsend
}
func MailSendAddTo(mailsend *MailSend, address string, name string) {
	to := make(map[string]string)
	to["address"] = address
	to["name"] = name
	mailsend.to = append(mailsend.to, to)
}

func MailSendAddAddressBook(mailsend *MailSend, addressbook string) {
	mailsend.addressbook = append(mailsend.addressbook, addressbook)
}

func MailSendSetSender(mailsend *MailSend, from string, fromname string) {
	mailsend.from = from
	mailsend.fromname = fromname
}
func MailSendSetReply(mailsend *MailSend, reply string) {
	mailsend.reply = reply
}
func MailSendAddCc(mailsend *MailSend, address string, name string) {
	cc := make(map[string]string)
	cc["address"] = address
	cc["name"] = name
	mailsend.cc = append(mailsend.cc, cc)
}
func MailSendAddBcc(mailsend *MailSend, address string, name string) {
	bcc := make(map[string]string)
	bcc["address"] = address
	bcc["name"] = name
	mailsend.bcc = append(mailsend.bcc, bcc)
}
func MailSendSetSubject(mailsend *MailSend, subject string) {
	mailsend.subject = subject
}
func MailSendSetText(mailsend *MailSend, text string) {
	mailsend.text = text
}
func MailSendSetHtml(mailsend *MailSend, html string) {
	mailsend.html = html
}

func MailSendAddVar(mailsend *MailSend, key string, value string) {
	mailsend.vars[key] = value
}
func MailSendAddLink(mailsend *MailSend, key string, value string) {
	mailsend.links[key] = value
}
func MailSendAddHeader(mailsend *MailSend, key string, value string) {
	mailsend.headers[key] = value
}

func MailSendBuildRequest(mailsend *MailSend) map[string]string {
	request := make(map[string]string)
	if len(mailsend.to) != 0 {
		to_list := make([]string, 0, 32)
		for _, key := range mailsend.to {
			to_list = append(to_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["to"] = strings.Join(to_list, ",")
	}
	if len(mailsend.addressbook) != 0 {
		request["addressbook"] = strings.Join(mailsend.addressbook, ",")
	}
	if mailsend.from != "" {
		request["from"] = mailsend.from
	}
	if mailsend.fromname != "" {
		request["from_name"] = mailsend.fromname
	}
	if mailsend.reply != "" {
		request["reply"] = mailsend.reply
	}
	if len(mailsend.cc) != 0 {
		cc_list := make([]string, 0, 32)
		for _, key := range mailsend.cc {
			cc_list = append(cc_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["cc"] = strings.Join(cc_list, ",")
	}
	if len(mailsend.bcc) != 0 {
		bcc_list := make([]string, 0, 32)
		for _, key := range mailsend.bcc {
			bcc_list = append(bcc_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["bcc"] = strings.Join(bcc_list, ",")
	}
	if mailsend.subject != "" {
		request["subject"] = mailsend.subject
	}
	if mailsend.text != "" {
		request["text"] = mailsend.text
	}
	if mailsend.html != "" {
		request["html"] = mailsend.html
	}
	if len(mailsend.vars) != 0 {
		data, err := json.Marshal(mailsend.vars)
		if err == nil {
			request["vars"] = string(data)
		}
	}
	if len(mailsend.links) != 0 {
		data, err := json.Marshal(mailsend.links)
		if err == nil {
			request["links"] = string(data)
		}
	}
	if len(mailsend.headers) != 0 {
		data, err := json.Marshal(mailsend.headers)
		if err == nil {
			request["headers"] = string(data)
		}
	}
	return request
}

//mailxsend
type MailXSend struct {
	to          []map[string]string
	addressbook []string
	from        string
	fromname    string
	reply       string
	cc          []map[string]string
	bcc         []map[string]string
	subject     string
	project     string
	vars        map[string]string
	links       map[string]string
	headers     map[string]string
}

func CreateMailXSend() *MailXSend {
	mailxsend := new(MailXSend)
	mailxsend.vars = make(map[string]string)
	mailxsend.links = make(map[string]string)
	mailxsend.headers = make(map[string]string)
	return mailxsend
}
func MailXSendAddTo(mailxsend *MailXSend, address string, name string) {
	to := make(map[string]string)
	to["address"] = address
	to["name"] = name
	mailxsend.to = append(mailxsend.to, to)
}
func MailXSendAddAddressBook(mailxsend *MailXSend, addressbook string) {
	mailxsend.addressbook = append(mailxsend.addressbook, addressbook)
}
func MailXSendSetSender(mailxsend *MailXSend, from string, fromname string) {
	mailxsend.from = from
	mailxsend.fromname = fromname
}
func MailXSendSetReply(mailxsend *MailXSend, reply string) {
	mailxsend.reply = reply
}
func MailXSendAddCc(mailxsend *MailXSend, address string, name string) {
	cc := make(map[string]string)
	cc["address"] = address
	cc["name"] = name
	mailxsend.cc = append(mailxsend.cc, cc)
}
func MailXSendAddBcc(mailxsend *MailXSend, address string, name string) {
	bcc := make(map[string]string)
	bcc["address"] = address
	bcc["name"] = name
	mailxsend.bcc = append(mailxsend.bcc, bcc)
}
func MailXSendSetSubject(mailxsend *MailXSend, subject string) {
	mailxsend.subject = subject
}
func MailXSendSetProject(mailxsend *MailXSend, project string) {
	mailxsend.project = project
}

func MailXSendAddVar(mailxsend *MailXSend, key string, value string) {
	mailxsend.vars[key] = value
}
func MailXSendAddLink(mailxsend *MailXSend, key string, value string) {
	mailxsend.links[key] = value
}
func MailXSendAddHeader(mailxsend *MailXSend, key string, value string) {
	mailxsend.headers[key] = value
}

func MailXSendBuildRequest(mailxsend *MailXSend) map[string]string {
	request := make(map[string]string)
	if len(mailxsend.to) != 0 {
		to_list := make([]string, 0, 32)
		for _, key := range mailxsend.to {
			to_list = append(to_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["to"] = strings.Join(to_list, ",")
	}
	if len(mailxsend.addressbook) != 0 {
		request["addressbook"] = strings.Join(mailxsend.addressbook, ",")
	}
	if mailxsend.from != "" {
		request["from"] = mailxsend.from
	}
	if mailxsend.fromname != "" {
		request["from_name"] = mailxsend.fromname
	}
	if mailxsend.reply != "" {
		request["reply"] = mailxsend.reply
	}
	if len(mailxsend.cc) != 0 {
		cc_list := make([]string, 0, 32)
		for _, key := range mailxsend.cc {
			cc_list = append(cc_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["cc"] = strings.Join(cc_list, ",")
	}
	if len(mailxsend.bcc) != 0 {
		bcc_list := make([]string, 0, 32)
		for _, key := range mailxsend.bcc {
			bcc_list = append(bcc_list, fmt.Sprintf("%s<%s>", key["name"], key["address"]))
		}
		request["bcc"] = strings.Join(bcc_list, ",")
	}
	if mailxsend.subject != "" {
		request["subject"] = mailxsend.subject
	}
	if mailxsend.project != "" {
		request["project"] = mailxsend.project
	}

	if len(mailxsend.vars) != 0 {
		data, err := json.Marshal(mailxsend.vars)
		if err == nil {
			request["vars"] = string(data)
		}
	}
	if len(mailxsend.links) != 0 {
		data, err := json.Marshal(mailxsend.links)
		if err == nil {
			request["links"] = string(data)
		}
	}
	if len(mailxsend.headers) != 0 {
		data, err := json.Marshal(mailxsend.headers)
		if err == nil {
			request["headers"] = string(data)
		}
	}
	return request
}

//messagexsend
type MessageXSend struct {
	to          []string
	addressbook []string
	project     string
	vars        map[string]string
}

func CreateMessageXSend() *MessageXSend {
	messagexsend := new(MessageXSend)
	messagexsend.vars = make(map[string]string)
	return messagexsend
}
func MessageXSendAddTo(messagexsend *MessageXSend, address string) {
	messagexsend.to = append(messagexsend.to, address)
}
func MessageXSendAddAddressBook(messagexsend *MessageXSend, addressbook string) {
	messagexsend.addressbook = append(messagexsend.addressbook, addressbook)
}

func MessageXSendSetProject(messagexsend *MessageXSend, project string) {
	messagexsend.project = project
}

func MessageXSendAddVar(messagexsend *MessageXSend, key string, value string) {
	messagexsend.vars[key] = value
}

func MessageXSendBuildRequest(messagexsend *MessageXSend) map[string]string {
	request := make(map[string]string)
	if len(messagexsend.to) != 0 {
		request["to"] = strings.Join(messagexsend.to, ",")
	}
	if len(messagexsend.addressbook) != 0 {
		request["addressbook"] = strings.Join(messagexsend.addressbook, ",")
	}

	if messagexsend.project != "" {
		request["project"] = messagexsend.project
	}

	if len(messagexsend.vars) != 0 {
		data, err := json.Marshal(messagexsend.vars)
		if err == nil {
			request["vars"] = string(data)
		}
	}
	return request
}

//addressbookmail
type AddressBookMail struct {
	address string
	target  string
}

func CreateAddressBookMail() *AddressBookMail {
	return new(AddressBookMail)
}
func AddressBookMailSetAddress(addressbookmail *AddressBookMail, address string, name string) {
	addressbookmail.address = fmt.Sprintf("%s<%s>", address, name)
}
func AddressBookMailSetAddressBook(addressbookmail *AddressBookMail, addressbook string) {
	addressbookmail.target = addressbook
}
func AddressBookMailBuildRequest(addressbookmail *AddressBookMail) map[string]string {
	request := make(map[string]string)
	request["address"] = addressbookmail.address
	if addressbookmail.target != "" {
		request["target"] = addressbookmail.target
	}
	return request
}

//addressbookmessage
type AddressBookMessage struct {
	address string
	target  string
}

func CreateAddressBookMessage() *AddressBookMessage {
	return new(AddressBookMessage)
}
func AddressBookMessageSetAddress(addressbookmessage *AddressBookMessage, address string, name string) {
	addressbookmessage.address = fmt.Sprintf("%s<%s>", address, name)
}
func AddressBookMessageSetAddressBook(addressbookmessage *AddressBookMessage, addressbook string) {
	addressbookmessage.target = addressbook
}
func AddressBookMessageBuildRequest(addressbookmessage *AddressBookMessage) map[string]string {
	request := make(map[string]string)
	request["address"] = addressbookmessage.address
	if addressbookmessage.target != "" {
		request["target"] = addressbookmessage.target
	}
	return request
}
