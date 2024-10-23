package lib

import (
	"fmt"
	"reflect"
)

type PodInfo struct {
	PODNAME          string
	IPADDRESS        string
	ENIID            string
	SECURITYGROUPIDS []string
}

func validatePodInfo(data interface{}) error {
	pods, ok := data.([]PodInfo)
	if !ok {
		return fmt.Errorf("invalid type, expected []PodInfo")
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods to output")
	}

	for _, pod := range pods {
		v := reflect.ValueOf(pod)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			switch field.Kind() {
			case reflect.String, reflect.Slice:
				// 文字列型とスライス型のフィールドはそのまま
			default:
				return fmt.Errorf("invalid type for field %s: %s", v.Type().Field(i).Name, field.Kind())
			}
		}
	}
	return nil
}
