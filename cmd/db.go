package cmd

import (
	"io/ioutil"
	"time"
	
	"gopkg.in/yaml.v2"
)

type database struct {
	Created time.Time    `yaml:"created"`
	Updated time.Time    `yaml:"updated"`
	Jails   []*jailProps `yaml:"jails"`
}

func loadDatabase() (*database, error) {
	var db database
	b, err := ioutil.ReadFile(CfgFile)
	err = yaml.Unmarshal(b, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *database) persist() error {
	if db.Created.IsZero() {
		db.Created = time.Now()
	}
	db.Updated = time.Now()
	d, err := yaml.Marshal(db)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(CfgFile, d, 0)
}

func (db *database) jailsYaml() (string, error) {
	d, err := yaml.Marshal(db.Jails)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
