package lib

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

func OutputTable(pods []PodInfo, out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	var header []string
	v := reflect.TypeOf(PodInfo{})
	var headerFormat string
	for i := 0; i < v.NumField(); i++ {
		header = append(header, v.Field(i).Name)
		headerFormat += "%s\t"
	}
	headerInterface := make([]interface{}, len(header))
	for i, v := range header {
		headerInterface[i] = v
	}
	fmt.Fprintf(w, headerFormat+"\n", headerInterface...)

	for _, pod := range pods {
		v := reflect.ValueOf(pod)
		var values []interface{}
		var format string
		for i := 0; i < v.NumField(); i++ {
			values = append(values, v.Field(i).Interface())
			format += "%v\t"
		}
		format = format[:len(format)-1]
		fmt.Fprintf(w, format+"\n", values...)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Failed to flush tabwriter: %v", err)
	}
	return nil
}
