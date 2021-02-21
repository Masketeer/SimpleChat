package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Server struct {
	Port  int32   `json:"Port"`
}

var ServerConfig Server

func init() {
	data, err := ioutil.ReadFile("conf.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = json.Unmarshal(data, &ServerConfig)
	if err != nil {
		fmt.Println(err.Error())
	}
}