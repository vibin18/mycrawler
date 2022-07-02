package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type UserData interface {
	GetUsers()
}

type fileUser struct {
	Id     int64  `yaml:"id"`
	Name   string `yaml:"name"`
	Number int64  `yaml:"number"`
}

type fileData struct {
	filename string
}

type fileUsers struct {
	Users []fileUser `yaml:"Users"`
}

func (u *fileData) GetUsers() *fileUsers {
	var users *fileUsers
	yamlFile, err := ioutil.ReadFile("user.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &users)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return users
}
