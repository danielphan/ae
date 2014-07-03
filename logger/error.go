package logger

import (
	"appengine"
	"encoding/json"
)

type errorLog struct {
	Log
	Message string
}

func Error(c appengine.Context, e error) string {
	el := errorLog{
		Log: Log{Type: "Error"},
		Message: e.Error(),
	}

	var s string
	b, err := json.Marshal(el)
	if err != nil {
		s = `{"Type":"Error","Message":"could not serialize error"}`
		c.Criticalf("%s", e)
		c.Criticalf("%s", err)
	} else {
		s = string(b)
	}
	c.Errorf("%s", s)
	return s
}
