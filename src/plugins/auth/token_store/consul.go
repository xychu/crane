package token_store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dataman-Cloud/borgsphere/src/pkg/auth"
	"github.com/Dataman-Cloud/borgsphere/src/util/consulclient"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

const (
	EXPIRATION_CHECK_INTERVAL = time.Minute * 5
	TOKEN_PREFIX              = "borgsphere_token_"
)

type consulTokenEntry struct {
	Token     string
	AccountId string
	ExpireAt  time.Time
}

type ConsulStore struct {
	auth.TokenStore
	Client   *consulclient.Client
	StopChan chan bool
}

// NewConsulStore returns a new ConsulStore
// the default consul server is http://127.0.0.1:8500
// also you can add env CONSUL_HTTP_ADDR customized
func NewConsulStore() *ConsulStore {
	client, err := consulclient.NewClient(consulclient.DefaultConfig())
	if err != nil {
		panic("new consul client error: " + err.Error())
	}
	cs := &ConsulStore{
		Client:   client,
		StopChan: make(chan bool, 1),
	}
	go cs.StartCheckExpire()
	return cs
}

// Set is used to save kvPair
func (c *ConsulStore) Set(ctx *gin.Context, token, accountId string, expiredAt time.Time) error {
	log.Debugf("Set %s %s %v", token, accountId, expiredAt)
	b, err := json.Marshal(consulTokenEntry{
		Token:     c.formattedToken(token),
		AccountId: accountId,
		ExpireAt:  expiredAt})
	if err != nil {
		log.Error("marshal consul consulTokenEntry error: ", err)
		return err
	}

	p := &consulclient.KVPair{Key: c.formattedToken(token), Value: b}
	c.Client.KV().Put(p, nil)
	return nil
}

// Get is used to lookup a userId with token
func (c *ConsulStore) Get(ctx *gin.Context, token string) (string, error) {
	log.Debug("Get ", token)
	kvPair, _, err := c.Client.KV().Get(c.formattedToken(token), nil)
	if err != nil {
		log.Errorf("get kvPair with key %s from consul error: %s", c.formattedToken(token), err.Error())
		return "", err
	}
	if kvPair == nil {
		return "", nil
	}

	var t consulTokenEntry
	if err = json.Unmarshal(kvPair.Value, &t); err != nil {
		log.Error("Unmarshal kvPair error: ", err)
		return "", err
	}

	return t.AccountId, nil
}

// Delete is used to delete a token
func (c *ConsulStore) Del(ctx *gin.Context, token string) error {
	log.Debug("del ", token)
	if _, err := c.Client.KV().Delete(c.formattedToken(token), nil); err != nil {
		log.Error("delete from consul error: ", err)
		return err
	}
	return nil
}

//List returns all the tokens
func (c *ConsulStore) List(token_prefix string) ([]*consulTokenEntry, error) {
	log.Debug("list ", token_prefix)
	kvPairs, _, err := c.Client.KV().List(token_prefix, nil)
	if err != nil {
		log.Error("list kvPairs from consul error: ", err)
		return nil, err
	}

	var ts []*consulTokenEntry
	for _, kvPair := range kvPairs {
		var t consulTokenEntry
		if err = json.Unmarshal(kvPair.Value, &t); err != nil {
			log.Error("Unmarshal kvPair error: ", err)
			return nil, err
		}
		ts = append(ts, &t)
	}
	return ts, nil
}

// start to check the token expiration , delete it if the token is out of date
func (c *ConsulStore) StartCheckExpire() {
	log.Info("start to check token expiration")
	for {
		select {
		case <-time.Tick(EXPIRATION_CHECK_INTERVAL):
			// check expiration
			ts, err := c.List(TOKEN_PREFIX)
			if err != nil {
				log.Error("list token error: ", err)
			}

			for _, t := range ts {
				if time.Now().Unix() > t.ExpireAt.Unix() {
					log.Debug("del ", t.Token)
					c.Del(new(gin.Context), t.Token)
				}
			}
		case <-c.StopChan:
			return
		}
	}
}

// stop to check token expiration
func (c *ConsulStore) StopCheckExpire() {
	log.Info("stop check expire")
	c.StopChan <- true
}

// format token with token prefix
func (c *ConsulStore) formattedToken(token string) string {
	return fmt.Sprintf("%s%s", TOKEN_PREFIX, token)
}
