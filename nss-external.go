package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	nss "github.com/protosam/go-libnss"
	nssStructs "github.com/protosam/go-libnss/structs"
)

// Placeholder main() stub is neccessary for compile.
func main() {}

func init() {
	// We set our implementation to "LibNssEtcd", so that go-libnss will use the methods we create
	nss.SetImpl(LibNssExternal{})
}

// LibNssExternal creates a struct that implements LIBNSS stub methods.
type LibNssExternal struct{ nss.LIBNSS }

type NSSConfig struct {
	Home    string
	MinUID  uint `yaml:"minuid"`
	GroupIP uint `yaml:"groupid"`
	Bash    string
	Suffix  []string
}

// Config is base config in /etc/nss_external.conf
type Config struct {
	Users []nssStructs.Passwd
	Nss   NSSConfig
}

var nsscfg *Config

func loadConfig() *Config {
	if nsscfg != nil {
		return nsscfg
	}
	cfgfile, cfgErr := ioutil.ReadFile("/etc/nss_external.conf")
	if cfgErr != nil {
		cfg := Config{}
		cfg.Nss = NSSConfig{}
		cfg.Nss.GroupIP = 1000
		cfg.Nss.MinUID = 10000
		cfg.Nss.Home = "/home/external/%s"
		cfg.Nss.Bash = "/bin/bash"
		cfg.Users = make([]nssStructs.Passwd, 0)
		nsscfg = &cfg
		return nsscfg
	}
	config := Config{}
	yaml.Unmarshal([]byte(cfgfile), &config)
	nsscfg = &config
	return nsscfg
}

////////////////////////////////////////////////////////////////
// Passwd Methods
////////////////////////////////////////////////////////////////

// PasswdAll will populate all entries for libnss
func (self LibNssExternal) PasswdAll() (nss.Status, []nssStructs.Passwd) {
	return nss.StatusSuccess, []nssStructs.Passwd{}
}

// PasswdByName returns a single entry by name.
func (self LibNssExternal) PasswdByName(name string) (nss.Status, nssStructs.Passwd) {
	// fmt.Printf("PasswdByName %s\n", name)
	cfg := loadConfig()
	// Accept only for usernames ending with @XXX XXX defined in config
	for _, suffix := range cfg.Nss.Suffix {
		if !strings.HasSuffix(name, suffix) {
			return nss.StatusNotfound, nssStructs.Passwd{}
		}
	}

	minUID := cfg.Nss.MinUID
	for _, user := range cfg.Users {
		if user.Username == name {
			return nss.StatusSuccess, user
		}
		if user.UID > minUID {
			minUID = user.UID
		}
	}
	fmt.Printf("User doesn't exists, add it")
	passwd := nssStructs.Passwd{
		Username: fmt.Sprintf("%s", name),
		Password: "*",
		UID:      minUID + 1,
		GID:      cfg.Nss.GroupIP,
		Shell:    cfg.Nss.Bash,
		Dir:      fmt.Sprintf(cfg.Nss.Home, name),
		Gecos:    fmt.Sprintf("external user %s", name),
	}
	cfg.Users = append(cfg.Users, passwd)
	// fmt.Printf("??%+v\n", cfg)
	newcfg, _ := yaml.Marshal(cfg)
	ioutil.WriteFile("/etc/nss_external.conf", newcfg, 0755)
	if _, err := os.Stat(fmt.Sprintf(cfg.Nss.Home, name)); os.IsNotExist(err) {
		err = os.MkdirAll(fmt.Sprintf(cfg.Nss.Home, name), 0775)
		if err != nil {
			fmt.Printf("failed to create home directory %s", err)
		}
		os.Chown(fmt.Sprintf(cfg.Nss.Home, name), int(passwd.UID), int(passwd.GID))
	}
	return nss.StatusSuccess, passwd
	//return nss.StatusNotfound, nssStructs.Passwd{}
}

// PasswdByUid returns a single entry by uid.
func (self LibNssExternal) PasswdByUid(uid uint) (nss.Status, nssStructs.Passwd) {
	// fmt.Printf("PasswdByUid %d skip\n", uid)
	cfg := loadConfig()
	for _, user := range cfg.Users {
		// fmt.Printf("search uid %d =? %d", user.UID, uid)
		if user.UID == uid {
			// fmt.Printf("User= %+v\n", user)
			return nss.StatusSuccess, user
		}
	}
	return nss.StatusNotfound, nssStructs.Passwd{}

}

// GroupAll returns all groups, not managed here
func (self LibNssExternal) GroupAll() (nss.Status, []nssStructs.Group) {
	// fmt.Printf("GroupAll\n")
	return nss.StatusSuccess, []nssStructs.Group{}
}

// GroupByName returns a group, not managed here
func (self LibNssExternal) GroupByName(name string) (nss.Status, nssStructs.Group) {
	// fmt.Printf("GroupByName %s\n", name)
	return nss.StatusNotfound, nssStructs.Group{}
}

// GroupBuGid retusn group by id, not managed heresSS
func (self LibNssExternal) GroupByGid(gid uint) (nss.Status, nssStructs.Group) {
	// fmt.Printf("GroupByGid %d\n", gid)
	return nss.StatusNotfound, nssStructs.Group{}
}

////////////////////////////////////////////////////////////////
// Shadow Methods
////////////////////////////////////////////////////////////////
// ShadowAll return all shadow entries, not managed as no password are allowed here
func (self LibNssExternal) ShadowAll() (nss.Status, []nssStructs.Shadow) {
	// fmt.Printf("ShadowAll\n")
	return nss.StatusSuccess, []nssStructs.Shadow{}
}

// ShadowByName return shadow entry, not managed as no password are allowed here
func (self LibNssExternal) ShadowByName(name string) (nss.Status, nssStructs.Shadow) {
	// fmt.Printf("ShadowByName %s\n", name)
	return nss.StatusNotfound, nssStructs.Shadow{}
}
