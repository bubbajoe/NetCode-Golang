package netcode

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func find(a go.Collection, b bson.M) interface{} {
    c := interface{}
    a.Find(a).One(&c)
    return c
}