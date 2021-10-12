package DeploymentTemplateValidate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/devtron-labs/devtron/pkg/pipeline"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/xeipuuv/gojsonschema"
)

type (
	UnitChecker   struct{}
	MemoryChecker struct{}
)

var (
	UChecker, _   = regexp.Compile("^([0-9.]+)m$")
	NoUChecker, _ = regexp.Compile("^([0-9.]+)$")
	MiChecker, _  = regexp.Compile("^[0-9]+Mi$")
	GiChecker, _  = regexp.Compile("^[0-9]+Gi$")
	TiChecker, _  = regexp.Compile("^[0-9]+Ti$")
	PiChecker, _  = regexp.Compile("^[0-9]+Pi$")
	KiChecker, _  = regexp.Compile("^[0-9]+Ki$")
)

func (f UnitChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return true
	}

	if UChecker.MatchString(asString) {
		return true
	} else if NoUChecker.MatchString(asString) {
		return true
	} else {
		return false
	}
}

func (f MemoryChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return true
	}

	// fmt.Println("hello", asString)
	if MiChecker.MatchString(asString) {
		return true
	} else if GiChecker.MatchString(asString) {
		return true
	} else if TiChecker.MatchString(asString) {
		return true
	} else if PiChecker.MatchString(asString) {
		return true
	} else if KiChecker.MatchString(asString) {
		return true
	} else {
		return false
	}
}

const memoryPattern = `"100Mi" or "1Gi" or "1Ti"`
const cpuPattern = `"50m" or "0.05"`
const cpu = "cpu"
const memory = "memory"

func DeploymentTemplateValidate(templatejson pipeline.TemplateRequest, schemafile string) (bool, error) {
	jsonFile, _ := os.Open(fmt.Sprintf("schema/%s.json", schemafile))
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var schemajson map[string]interface{}
	json.Unmarshal([]byte(byteValue), &schemajson)
	schemaLoader := gojsonschema.NewGoLoader(schemajson)
	documentLoader := gojsonschema.NewGoLoader(templatejson)
	buff, err := json.Marshal(templatejson)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	if result.Valid() {
		var dat map[string]interface{}

		if err := json.Unmarshal(buff, &dat); err != nil {
			log.Fatal(err)
			return false, err
		}
		//limits and requests are mandatory fields in schema
		for _, i := range []string{"valuesOverride", "defaultAppOverride"} {
			autoscaleEnabled := dat[i].(map[string]interface{})["autoscaling"].(map[string]interface{})
			if autoscaleEnabled["enabled"].(bool) {
				limit := dat[i].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
				request := dat[i].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})

				cpu_limit, _ := util2.CpuToNumber(limit["cpu"].(string))
				memory_limit, _ := util2.MemoryToNumber(limit["memory"].(string))
				cpu_request, _ := util2.CpuToNumber(request["cpu"].(string))
				memory_request, _ := util2.MemoryToNumber(request["memory"].(string))

				envoproxy_limit := dat[i].(map[string]interface{})["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
				envoproxy_request := dat[i].(map[string]interface{})["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})

				envoproxy_cpu_limit, _ := util2.CpuToNumber(envoproxy_limit["cpu"].(string))
				envoproxy_memory_limit, _ := util2.MemoryToNumber(envoproxy_limit["memory"].(string))
				envoproxy_cpu_request, _ := util2.CpuToNumber(envoproxy_request["cpu"].(string))
				envoproxy_memory_request, _ := util2.MemoryToNumber(envoproxy_request["memory"].(string))
				if (envoproxy_cpu_limit < envoproxy_cpu_request) || (envoproxy_memory_limit < envoproxy_memory_request) || (cpu_limit < cpu_request) || (memory_limit < memory_request) {
					return false, errors.New("requests is greater than limits")
				}

			}
		}
		fmt.Println("ok")
		return true, nil
	} else {
		var stringerror string
		fmt.Printf("The document is not valid. see errors :\n")
		for _, err := range result.Errors() {
			fmt.Println(err.Details()["format"])
			if err.Details()["format"] == cpu {
				stringerror = stringerror + "Error in " + err.Field() + ". Format should be like " + cpuPattern + "\n"
			} else if err.Details()["format"] == memory {
				stringerror = stringerror + "Error in " + err.Field() + ". Format should be like " + memoryPattern + "\n"
			} else {
				stringerror = stringerror + err.String() + "\n"
			}
		}
		return false, errors.New(stringerror)
	}
}
