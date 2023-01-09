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

// init sets up config from the env as well as global vars
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

// main is the standard entry point into the application
func main() {
	log.Printf("Fetching a batch of patients")
	// fetching a bundle of patients.  Bundle is a FHIR specific Resource
	bundle, err := getPatients()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatalln("error fetching patient batch")
	}

	// loop the Bundle's Entries
	for _, e := range bundle.Entry {
		var p fhir.Patient
		_ = json.Unmarshal(e.Resource, &p)
		log.Printf("Fetching a single patient with an id of: (%s)", *p.Id)
		// grab a single patient by id.  Patient is a FHIR resource
		// the Healthlake API is REST so the ID makes the Resource
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

// getPatients fetches a FHIR Bundle based upon query parameters
// in the below implementation, there is nothing being passed in
func getPatients() (*fhir.Bundle, error) {
	url := fmt.Sprintf("https://%s/%s/r4/Patient", healthlakeEndpoint, healthlakeDatastore)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// the request must by V4 signed
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

// getPatientById fetches a FHIR Patient based upon the ID supplied
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
