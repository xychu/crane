package token_store

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestConsulStore(t *testing.T) {
	c := NewConsulStore()
	var token string = "555"
	var accountId string = "111"
	if err := c.Set(new(gin.Context), token, accountId, time.Now().Add(time.Minute*1)); err != nil {
		t.Error(err)
	} else {
		t.Log("test Set func successfully")
	}

	id, err := c.Get(new(gin.Context), token)
	if err != nil {
		t.Error(err)
	}

	if id == accountId {
		t.Log("test Get func successfully")
	} else {
		t.Log("test Get func failed")
	}

	if err = c.Del(new(gin.Context), token); err != nil {
		t.Error(err)
	} else {
		t.Log("test Del func successfully")
	}
}

func TestList(t *testing.T) {
	c := NewConsulStore()
	ts, err := c.List("borgsphere_token")
	if err != nil {
		t.Error(err)
	}
	for _, tt := range ts {
		t.Log("kv == ", tt.AccountId, tt.ExpireAt)
	}
	t.Log("test List func successfully")
}

func TestStartCheckExpire(t *testing.T) {
	c := NewConsulStore()
	go c.StartCheckExpire()
	c.StopCheckExpire()

	t.Log("test startExpire and stopExpire func successfully")
}
