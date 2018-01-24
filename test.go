package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID bson.ObjectId
	username string	
	firstname string
	lastname string	
	password string
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println("MongoDB Error: ", err)
		
	}
	userCollection := session.DB("Netcode").C("users")
	user := &User{

	}
	users := []User{
	User{
		username: "skb",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "bubba",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "joe",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "admin",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "mod",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "root",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "me",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "username",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		username: "user",
		password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	}}
	errr := userCollection.Insert(users)
	if errr != nil {
		fmt.Println("Insert Error: ", err)
		
	}
	q := userCollection.Find(bson.M{"username":"bubbajoe"})
	
	if user != nil {
		fmt.Println(user.ID.String())
		fmt.Println(user)
		fmt.Println(user.password)
	} else {
		fmt.Println("None")
	}
}