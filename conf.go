package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	dir        string
	rdb        []RDBSnapshot
	rdbFn      string
	aoffn      string
	aofFsync   FSyncMode
	aofEnabled bool
}

type RDBSnapshot struct {
	Secs        int
	KeysChanged int
}

type FSyncMode string

const (
	Always   FSyncMode = "always"
	EverySec FSyncMode = "everysec"
	No       FSyncMode = "no"
)

func NewConfig() *Config {
	return &Config{}
}

func readConf(fn string) *Config {
	conf := NewConfig()
	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("Cannot read %s - using default config\n", fn)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		parseLine(line, conf)

	}

	if err := s.Err(); err != nil {
		fmt.Println("error scanning config file: ", err)
		return conf
	}

	if conf.dir != "" {
		os.MkdirAll(conf.dir, 0755)
	}

	return conf
}

func parseLine(line string, conf *Config) {
	args := strings.Split(line, " ")
	cmd := args[0]
	switch cmd {
	case "save":
		secs, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("invalid secs")
			return
		}

		keysChanged, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("invalid keys count")
			return
		}

		snapshot := RDBSnapshot{
			Secs:        secs,
			KeysChanged: keysChanged,
		}
		conf.rdb = append(conf.rdb, snapshot)
	case "dbfilename":
		conf.rdbFn = args[1]
	case "appendfilename":
		conf.aoffn = args[1]
	case "appendfsync":
		conf.aofFsync = FSyncMode(args[1])
	case "dir":
		conf.dir = args[1]
	case "appendonly":
		if args[1] == "yes" {
			conf.aofEnabled = true
		} else {
			conf.aofEnabled = false
		}

	}

}
