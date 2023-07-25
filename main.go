package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"thallium/ping"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type config struct {
	Timeout duration
	Period  duration
	Log     bool

	Mysql struct {
		Connect string
		Online  string
		Offline string
		Insert  string
	}

	Target []Target
}

var (
	conf config
	db   *sqlx.DB
)

func main() {
	configFile := flag.String("config", "config.toml", "path to the config file")
	flag.Parse()

	var err error
	if _, err = toml.DecodeFile(*configFile, &conf); err != nil {
		log.Fatalln(err)
	}

	if db, err = sqlx.Open("mysql", conf.Mysql.Connect); err == nil {
		db.SetConnMaxLifetime(2 * time.Minute)
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(2)
	} else {
		log.Fatalln(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalln("mysql:", err)
	}

	for _, target := range conf.Target {
		go targetRoutine(target)
	}

	time.Sleep(time.Duration(1<<63 - 1))
}

func targetRoutine(target Target) {
	if target.Period.Duration == 0 {
		target.Period = conf.Period
	}
	tick := time.Tick(target.Period.Duration)
	adapter := &DbStatusAdapter{
		Name:    target.Name,
		Address: target.Address,
	}
	if _, err := db.NamedExec(conf.Mysql.Insert, adapter); err != nil {
		log.Println("mysql:", err)
	}
	var pinger ping.Pinger
	if target.Legacy {
		pinger = &ping.Legacy{
			Addr:         target.Address,
			LocalAddress: target.LocalAddress,
			Timeout:      conf.Timeout.Duration,
		}
	} else {
		pinger = &ping.Standard{
			Addr:         target.Address,
			LocalAddress: target.LocalAddress,
			Timeout:      conf.Timeout.Duration,
		}
	}
	for ; true; <-tick {
		status, pingError := pinger.Ping()
		query := ""
		if pingError != nil {
			query = conf.Mysql.Offline
			adapter.Online = 0
			adapter.Max = 0
			adapter.Latency = 0
			adapter.Favicon = ""
			adapter.Protocol = 0
			adapter.Version = ""
			adapter.Motd = ""
		} else {
			query = conf.Mysql.Online
			adapter.Online = status.Players.Online
			adapter.Max = status.Players.Max
			adapter.Latency = int(status.Latency.Milliseconds())
			adapter.Favicon = string(status.Favicon)
			adapter.Protocol = status.Version.Protocol
			adapter.Version = status.Version.Name
			if strings.Contains(query, ":motd") {
				adapter.Motd = status.Description.ClearString()
			}
		}
		adapter.Time = time.Now().Unix()
		if _, err := db.NamedExec(query, adapter); err != nil {
			log.Println(target.Name, "mysql:", err)
		} else if conf.Log {
			if pingError != nil {
				log.Printf(
					"%s: Error: %v",
					target.Name, pingError,
				)
			} else {
				log.Printf(
					"%s: Ping: %dms, Version: %s, Online: %d/%d",
					target.Name, adapter.Latency, adapter.Version, adapter.Online, adapter.Max,
				)
			}
		}
	}
}
