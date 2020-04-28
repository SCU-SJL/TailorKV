package tailor_server

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

type TailorConfig struct {
	XMLName           xml.Name `xml:"config"`
	DefaultExpiration string   `xml:"defaultExpiration"`
	CleanCycle        string   `xml:"cleanCycle"`
	AsyncCleanCycle   string   `xml:"asyncCleanCycle"`
	Concurrency       string   `xml:"concurrency"`
	SavingPath        string   `xml:"savingPath"`
	FileName          string   `xml:"fileName"`
}

func GetConfig(path string) *TailorConfig {
	file, err := os.Open(path) // For read access.
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	tailorConf := TailorConfig{}
	err = xml.Unmarshal(data, &tailorConf)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	return &tailorConf
}
