package deploy

import (
	"errors"
	"strings"
	"fmt"	
	"bytes"
	"encoding/xml"
	"github.com/adriaandejonge/xld/repo"
	"github.com/clbanning/mxj/j2x"		
	 "github.com/adriaandejonge/xld/util/intf"
	"github.com/adriaandejonge/xld/util/http"

)

func Do(args intf.Command) (result string, err error) {
	subs := args.Subs()
	if len(subs) < 0 {
		// TODO meaningful
		return "error", errors.New("xld deploy expects at least 0 arguments")
	} else {

		switch args.Main() {

		case "plan":
			return plan(args)

		case "deploy":
			return deploy(args)

		case "upgrade":
			return "error", errors.New("Update is not yet implemented")

		case "undeploy":
			return undeploy(args)

		default:
			return "error", errors.New("Unknown command")
		}
	}
}

type (
	Task struct {
		Id string `xml:"id,attr"`
		CurrentStep int `xml:"currentStep,attr"`
		TotalSteps int `xml:"totalSteps,attr"`
		Failures int `xml:"failures,attr"`
		State string `xml:"state,attr"`
		State2 string `xml:"state2,attr"`
		Owner string `xml:"owner,attr"`
		Description string `xml:"description"`
		Current int `xml:"currentSteps>current"`
		Environment string `xml:"metadata>environment"`
		TaskType string `xml:"metadata>taskType"`
		EnvironmentId string `xml:"metadata>environment_id"`
		Application string `xml:"metadata>application"`
		Version string `xml:"metadata>version"`
		Steps []Step `xml:"steps>step"`
	}

	Step struct {
		Failures int `xml:"failures,attr"`
		State string `xml:"state,attr"`
		Description string `xml:"description"`
		Log string `xml:"log"`
		PreviewAvailable string `xml:"metadata>previewAvailable"`
		Order int `xml:"metadata>order"`
	}
)

func plan(args intf.Command) (result string, err error) {
	result, err = prepare(args)

	// TODO Read err

	_, body, err := http.Read("/task/" + result+ "/step")

	task := Task{}
	err = xml.Unmarshal(body, &task)
	if err != nil {
		return "error", err
	}

	fmt.Println("Plan", task.Description)

	for i, step := range task.Steps {
		fmt.Printf("%d/%d - " + step.Description + "\n", i + 1, task.TotalSteps)
	}



	return "", err
}

func deploy(args intf.Command) (result string, err error) {
	result, err = prepare(args)

	body, err := http.Create("/task/" + result+ "/start", nil)

	// TODO Read status and print updates

	return string(body), err
}

func undeploy(args intf.Command) (result string, err error) {
	subs := args.Subs()
	appToUndeploy := repo.AntiAbbreviate(subs[0])
	_, body, err := http.Read("/deployment/prepare/undeploy?deployedApplication=" + appToUndeploy)

	body, err = http.Create("/deployment", bytes.NewBuffer(body))

	body, err = http.Create("/task/" + string(body) + "/start", nil)

	// TODO Read status and print updates

	return string(body), err
}


func prepare(args intf.Command) (result string, err error) {
	subs := args.Subs()
	appVersion := repo.AntiAbbreviate(subs[0])
	targetEnv := repo.AntiAbbreviate(subs[1]) // or 2?

	parts := strings.Split(appVersion, "/")

	app := parts[len(parts) - 2]
	//version := parts[len(parts) - 1]

	deployment := map[string]interface{} {
		"deployment": map[string]interface{} {
			"-type": "INITIAL",
			"application": map[string]interface{} {
				"udm.DeployedApplication": map[string]interface{} {
					"-id": targetEnv + "/" + app,
					"version": map[string]interface{} {
						"-ref": appVersion,
					},
					"environment": map[string]interface{} {
						"-ref": targetEnv,
					},
					"optimizePlan": "true",
				},
			},
		},
	}


	// TODO Make this a util?
	json, _ := j2x.MapToJson(deployment)
	xml, _ := j2x.JsonToXml(json)


	body, err := http.Create("/deployment/prepare/deployeds", bytes.NewBuffer(xml))	
	if err != nil {
		return "error", err
	}

	body, err = http.Create("/deployment", bytes.NewBuffer(body))

	return string(body), err
}

