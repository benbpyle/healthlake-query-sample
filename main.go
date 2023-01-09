package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"

	"net/http"
	"os"
	"time"
)

var (
	healthlakeEndpoint  string
	healthlakeDatastore string
	httpClient          *http.Client
	signer              *v4.Signer
)

func init() {
	healthlakeEndpoint = os.Getenv("ENDPOINT")
	healthlakeDatastore = os.Getenv("DATASTORE")
	httpClient = &http.Client{}
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	signer = v4.NewSigner(sess.Config.Credentials)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
}

func main() {
	log.Printf("Fetching a batch of patients")
	bundle, err := getPatients()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatalln("error fetching patient batch")
	}

	for _, e := range bundle.Entry {
		var p fhir.Patient
		_ = json.Unmarshal(e.Resource, &p)
		log.Printf("Fetching a single patient with an id of: (%s)", *p.Id)
		patient, err := getPatientById(*p.Id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatalln("error fetching single patient")
		}

		log.WithFields(log.Fields{
			"patient": patient,
		}).Infof("printing out the patient")

	}

}

func getPatients() (*fhir.Bundle, error) {
	url := fmt.Sprintf("https://%s/%s/r4/Patient", healthlakeEndpoint, healthlakeDatastore)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, _ = signer.Sign(req, nil, "healthlake", "us-west-2", time.Now())

	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var bundle fhir.Bundle
	decoder := json.NewDecoder(resp.Body)
	//	b, err := io.ReadAll(resp.Body)
	err = decoder.Decode(&bundle)

	return &bundle, err
}

func getPatientById(id string) (*fhir.Patient, error) {
	url := fmt.Sprintf("https://%s/%s/r4/Patient/%s", healthlakeEndpoint, healthlakeDatastore, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	_, _ = signer.Sign(req, nil, "healthlake", "us-west-2", time.Now())

	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var patient fhir.Patient
	decoder := json.NewDecoder(resp.Body)
	//	b, err := io.ReadAll(resp.Body)
	err = decoder.Decode(&patient)

	return &patient, err
}
