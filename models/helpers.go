package models

import "gopkg.in/mgo.v2/bson"

//M is an alias for bson.M
type M bson.M

// NotFoundString is the string returned when an element is not found
var NotFoundString = "not found"
