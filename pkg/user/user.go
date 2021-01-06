package user

import (
	"fmt"

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

func GetUser(name string) (*User, error) {
	//TODO このviper.getの部分，cmd/secretary-lab/app/run.goに依存してるのよくない
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
