package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID bson.ObjectId 	`bson:"_id,omitempty"`
	Username string		`bson:"username,omitempty"`
	Firstname string	`bson:"firstname,omitempty"`
	Lastname string		`bson:"lastname,omitempty"`
	Password string		`bson:"password,omitempty"`
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println("MongoDB Error: ", err)
		
	}
	userCollection := session.DB("Netcode").C("users")
	user := &User{}
	users := []User{
	User{
		Username: "skb",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "bubba",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "joe",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "admin",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "mod",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "root",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "me",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "username",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	},User{
		Username: "user",
		Password: "$2a$04$17haVb1oGv0E4bQjZd5cvuRTWYiZFnVFpeNCNhttJ659rl/xDlrfq",
	}}
	errr := userCollection.Insert(users)
	if errr != nil {
		fmt.Println("Insert Error: ", err)
	}
	q := userCollection.Find(bson.M{"username":"bubbajoe"})
	q.One(&user)

	if user != nil {
		fmt.Println(user.ID.String())
		fmt.Println(user)
	} else {
		fmt.Println("None")
	}
}