package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Host struct {
	Hostname   string    `json:"hostname"`
	Expiration time.Time `json:"expiration"`
	Protocol   string    `json:"protocol"`
	Port       string    `json:"port"`
	Path       string    `json:"path"`
}

type Status struct {
	Hosts  []Host  `json:"hosts"`
	Uptime float64 `json:"uptime"`
}

type TravisNotification struct {
	Payload TravisPayload `json:"payload"`
}

type TravisPayload struct {
	Status     string           `json:"status_message"`
	Commit     string           `json:"commit"`
	Branch     string           `json:"branch"`
	Message    string           `json:"message"`
	Repository TravisRepository `json:"repository"`
}

type TravisRepository struct {
	Name  string `json:"name"`
	Owner string `json:"owner_name"`
}

var (
	hosts []Host
	TTL   = 30
)

func buildExpiration() time.Time {
	expiration := time.Now().UTC()
	expiration = expiration.Add(time.Duration(TTL) * time.Second)

	return expiration
}

func cleanup() {
	for {
		for i, h := range hosts {
			if time.Now().UTC().After(h.Expiration) {
				diff := hosts
				diff = append(diff[:i], diff[i+1:]...)
				hosts = diff
			}
		}
		time.Sleep(time.Duration(TTL) * time.Second)
	}
}

func main() {
	boot := time.Now()

	go cleanup()

	router := gin.Default()

	router.POST("/checkin/", func(c *gin.Context) {
		var host Host

		c.Bind(&host)

		if host.Hostname != "" {
			for i, h := range hosts {
				if h.Hostname == host.Hostname {
					diff := hosts
					diff = append(diff[:i], diff[i+1:]...)
					hosts = diff
				}
			}

			host.Expiration = buildExpiration()
			hosts = append(hosts, host)

			c.JSON(200, "")
		} else {
			var response struct {
				Message string `json:"message"`
			}
			response.Message = "Invalid hostname."
			c.JSON(400, response)
		}
	})

	router.POST("/announce/", func(c *gin.Context) {
		var notification TravisNotification

		c.Bind(&notification)

		payload, _ := json.Marshal(notification.Payload)
		body := bytes.NewBuffer(payload)

		for _, host := range hosts {
			req, err := http.NewRequest("POST", host.Hostname, body)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()
		}

		c.JSON(200, "")
	})

	router.GET("/status/", func(c *gin.Context) {
		status := Status{}
		status.Hosts = hosts
		status.Uptime = time.Now().Sub(boot).Seconds()

		c.JSON(200, status)
	})

	router.Run(":8080")
}
