package awsconfig

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type AWSProfile struct {
	Name           string
	Region         string
	HintEKSRegions []string
}

type AWSConfig struct {
	Profiles map[string]AWSProfile
}

var ConfigData *AWSConfig = nil

func loadAWSConfig() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error retrieving UserHomeDir")
		os.Exit(1)
	}

	file, err := os.Open(homeDir + "/.aws/config")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	awsConfig := AWSConfig{Profiles: make(map[string]AWSProfile)}
	awsConfig.Profiles[""] = AWSProfile{}

	// Read the file
	scanner := bufio.NewScanner(file)
	currentProfile := ""
	for scanner.Scan() {
		line := scanner.Text()

		// remove leading and trailing spaces
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "# kubectl-eks-regions=") {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				continue
			}
			eksRegions := strings.Split(parts[1], ",")

			for i, region := range eksRegions {
				eksRegions[i] = strings.TrimSpace(region)
			}

			currentProfileDetails := awsConfig.Profiles[currentProfile]
			currentProfileDetails.HintEKSRegions = eksRegions
			awsConfig.Profiles[currentProfile] = currentProfileDetails
		}

		// get rid of comments
		if strings.Contains(line, "#") {
			line = strings.Split(line, "#")[0]
		}

		// remove leading and trailing spaces
		line = strings.TrimSpace(line)

		// skip empty lines
		if len(line) == 0 {
			continue
		}

		// [profile example]
		if strings.HasPrefix(line, "[profile ") {
			currentProfile = strings.TrimSuffix(strings.TrimPrefix(line, "[profile "), "]")
			awsConfig.Profiles[currentProfile] = AWSProfile{Name: currentProfile}
		}

		parts := strings.Split(line, "=")

		// skip lines that are not key=value
		if len(parts) != 2 {
			continue
		}

		if strings.TrimSpace(parts[0]) == "region" {
			currentProfileDetails := awsConfig.Profiles[currentProfile]
			currentProfileDetails.Region = strings.TrimSpace(parts[1])
			awsConfig.Profiles[currentProfile] = currentProfileDetails
		}
	}

	ConfigData = &awsConfig
}

func GetAWSProfilesWithEKSHints() []AWSProfile {
	if ConfigData == nil {
		loadAWSConfig()
	}

	fmt.Printf("ConfigData: %+v\n", ConfigData)

	profiles := []AWSProfile{}
	for _, profileDetails := range ConfigData.Profiles {
		if len(profileDetails.HintEKSRegions) > 0 {
			profiles = append(profiles, profileDetails)
		}
	}

	return profiles
}
