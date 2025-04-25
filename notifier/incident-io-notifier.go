package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type IncidentIONotifier struct {
	Enabled     bool
	ClusterName string            `json:"cluster-name"`
	BaseURL     string            `json:"base-url"`
	Endpoint    string            `json:"endpoint"`
	Payload     map[string]string `json:"payload"`
	Metadata    map[string]string `json:"metadata"`
}

// NotifierName provides name for notifier selection
func (notifier *IncidentIONotifier) NotifierName() string {
	return "incidentio"
}

func (notifier *IncidentIONotifier) Copy() Notifier {
	n := *notifier
	return &n
}

// Notify sends messages to the endpoint notifier
func (notifier *IncidentIONotifier) Notify(messages Messages) bool {
	overallStatus, pass, warn, fail := messages.Summary()
	t := TemplateData{
		ClusterName:  notifier.ClusterName,
		SystemStatus: overallStatus,
		FailCount:    fail,
		WarnCount:    warn,
		PassCount:    pass,
		Nodes:        mapByNodes(messages),
	}
	values := map[string]any{}

	for key, val := range notifier.Payload {
		data, err := renderTemplate(t, "", val)
		if err != nil {
			log.Println("Error rendering template: ", err)
			return false
		}
		values[key] = string(data)
	}

	metadataValues := map[string]string{}
	for key, val := range notifier.Metadata {
		data, err := renderTemplate(t, "", val)
		if err != nil {
			log.Println("Error rendering template: ", err)
			return false
		}
		metadataValues[key] = string(data)
	}

	values["metadata"] = metadataValues

	requestBody, err := json.Marshal(values)
	if err != nil {
		log.Println("Unable to encode POST data")
		return false
	}

	endpoint := fmt.Sprintf("%s%s", notifier.BaseURL, notifier.Endpoint)
	if res, err := http.Post(endpoint, "application/json", bytes.NewBuffer(requestBody)); err != nil {
		log.Println("Unable to send data to incident.io endpoint:", err)
		return false
	} else {
		defer res.Body.Close()
		statusCode := res.StatusCode
		if statusCode != 202 {
			body, _ := ioutil.ReadAll(res.Body)
			log.Println("Unable to notify incident.io endpoint:", string(body))
			return false
		} else {
			log.Println("Notification sent to incident.io endpoint.")
			return true
		}
	}

}
