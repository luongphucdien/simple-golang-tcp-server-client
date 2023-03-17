package userinfo

import (
	"encoding/base64"
	"encoding/gob"
	"log"
	"os"
)

type User struct {
	Username  	string
	Password  	string
	FullName 	string
	Emails    	[]string
	Addresses 	[]string
}

func Encode(password string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(password))
	return enc
}

func decode(enc_password string) string {
	dec, err := base64.StdEncoding.DecodeString(enc_password)

	if err != nil {
		log.Fatal("Error in decoding: ", err)
	}

	return string(dec)
}

func saveUser(user_data any) {
	user_file, err := os.Create("user-info.SAVE")

	if err != nil {
		log.Fatal("Error in creating file: ", err)
	}

	enc := gob.NewEncoder(user_file)
	err = enc.Encode(user_data)

	if err != nil {
		log.Fatal("Error in encoding file: ", err)
	}
}

func LoadUser(path string) User{
	var user_obj User

	user_info, err := os.Open(path)

	if err != nil {
		log.Fatal("Error in reading file: ", err)
	}

	dec := gob.NewDecoder(user_info)
	err = dec.Decode(&user_obj)

	if err != nil {
		log.Fatal("Error in decoding file: ", err)
	}

	user_info.Close()

	return user_obj
}

func main() {
	user := User{
		"lpd",
		Encode("test"),
		"Luong Phuc Dien",
		[]string{"test-email@gmail.com"},
		[]string{"Earth"},
	}

	saveUser(user)

	// fmt.Print(LoadUser())
}
