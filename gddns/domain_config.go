package gddns

import (
	"crypto/rand"
	"encoding/base64"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type DomainConfig struct {
	Hostname  string
	Username  string
	Password  string
	DomainKey string
}


func ListDomains(c appengine.Context, u *user.User) ([]*DomainConfig, error) {
	userKey := domainUserKey(c, u)
	q := datastore.NewQuery("DomainConfig").Ancestor(userKey).Order("-Hostname").Limit(100)

	var domains []*DomainConfig
	if _, err := q.GetAll(c, &domains); err != nil {
		return nil, err
	}
	
	return domains, nil
}

func AddDomain(c appengine.Context, u *user.User, hostname string, username string, password string) (*DomainConfig, error) {
	dk, err := generateDomainKey()
	if err != nil {
		return nil, err
	}

	dc := DomainConfig{
		Hostname:  hostname,
		Username:  username,
		Password:  password,
		DomainKey: dk,
	}
	
	c.Infof("DomainConfig: %v", dc)

	userKey := domainUserKey(c, u)
	c.Infof("userKey: %v", userKey)
	
	dcKey := datastore.NewIncompleteKey(c, "DomainConfig", userKey)
	c.Infof("dcKey: %v", dcKey)
	
	if _, err := datastore.Put(c, dcKey, &dc); err != nil {
		c.Infof("err: %v", err)
		return nil, err
	}

	return &dc, nil
}

func GetDomain(c appengine.Context, domainKey string) (*DomainConfig, error) {
	q := datastore.NewQuery("DomainConfig").Filter("DomainKey =", domainKey)
	var domains []DomainConfig
	if _, err := q.GetAll(c, &domains); err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, nil
	}
	return &domains[0], nil
}

func domainUserKey(c appengine.Context, u *user.User) *datastore.Key {
	return datastore.NewKey(c, "DomainUser", u.ID, 0, nil)
}

func generateDomainKey() (string, error) {
	rb := make([]byte, 48)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(rb), nil
}
