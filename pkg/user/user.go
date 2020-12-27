package user

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

type User struct {
	Name                string
	SlackId             string
	SlackChannel        string
	SecretaryName       string
	SecretaryIcon       string
	SecretaryComingMsg  string
	SecretaryGoodbyeMsg string
}

func GetUser(configPath, name string) (*User, error) {
	path, fileName := filepath.Split(configPath)
	fileNameExt := filepath.Ext(fileName)
	fileName = fileName[0 : len(fileName)-len(fileNameExt)]

	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()

	if err != nil {
		return nil, err
	}
	users := viper.Get("users").(map[string]interface{})
	uInfo, ok := users[name].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("there is no user named %s", name)
	}

	u := User{
		Name:                name,
		SlackId:             uInfo["slack_id"].(string),
		SlackChannel:        uInfo["slack_channel"].(string),
		SecretaryName:       uInfo["secretary_name"].(string),
		SecretaryIcon:       uInfo["secretary_icon"].(string),
		SecretaryComingMsg:  uInfo["secretary_coming_msg"].(string),
		SecretaryGoodbyeMsg: uInfo["secretary_goodbye_msg"].(string),
	}
	return &u, nil
}
