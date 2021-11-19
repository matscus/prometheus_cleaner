package main

import (
	"bufio"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	configPath, prometheusHost, startDay string
)

func main() {
	flag.StringVar(&configPath, "configPath", "config.ini", "path to config file")
	flag.StringVar(&prometheusHost, "prometheusHost", "localhost:9090", "prometheus url")
	flag.StringVar(&startDay, "startday", "Sunday", "clean day")
	flag.Parse()
	for {
		day := time.Now().Weekday().String()
		if strings.ToLower(day) == strings.ToLower(startDay) {
			err := clean()
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Info("Not a start day, sleep 1 day")
			time.Sleep(24 * time.Hour)
		}
	}
}

func clean() error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Clean recover panic ", err)
		}
	}()
	file, err := os.Open(configPath)
	if err != nil {
		log.Panic(err)
	}

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		str := "http://" + prometheusHost + "/api/v1/admin/tsdb/delete_series?match[]={__name__=~\"" + string(line) + "\"}"
		req, err := http.NewRequest("POST", str, nil)
		if err != nil {
			return err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		log.Info("Delete rows ", string(line))
	}
	return nil
}
