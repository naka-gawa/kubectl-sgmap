package utils

import (
	"bytes"
	"testing"
)

func TestOutputJSON(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		{
			name: "VALID - single pod",
			args: args{
				data: []PodInfo{
					{
						PODNAME:          "pod1",
						IPADDRESS:        "192.168.1.1",
						ENIID:            "eni-1234567890abcdef",
						SECURITYGROUPIDS: []string{"sg-12345678"},
					},
				},
			},
			wantOut: `[
  {
    "PODNAME": "pod1",
    "IPADDRESS": "192.168.1.1",
    "ENIID": "eni-1234567890abcdef",
    "SECURITYGROUPIDS": [
      "sg-12345678"
    ]
  }
]`,
			wantErr: false,
		},
		{
			name: "VALID - multiple pods",
			args: args{
				data: []PodInfo{
					{
						PODNAME:          "pod1",
						IPADDRESS:        "192.168.1.1",
						ENIID:            "eni-1234567890abcdef",
						SECURITYGROUPIDS: []string{"sg-12345678"},
					},
					{
						PODNAME:          "pod2",
						IPADDRESS:        "192.168.1.2",
						ENIID:            "eni-abcdef1234567890",
						SECURITYGROUPIDS: []string{"sg-87654321"},
					},
				},
			},
			wantOut: `[
  {
    "PODNAME": "pod1",
    "IPADDRESS": "192.168.1.1",
    "ENIID": "eni-1234567890abcdef",
    "SECURITYGROUPIDS": [
      "sg-12345678"
    ]
  },
  {
    "PODNAME": "pod2",
    "IPADDRESS": "192.168.1.2",
    "ENIID": "eni-abcdef1234567890",
    "SECURITYGROUPIDS": [
      "sg-87654321"
    ]
  }
]`,
			wantErr: false,
		},
		{
			name: "INVALID - empty pods",
			args: args{
				data: []PodInfo{},
			},
			wantOut: "",
			wantErr: true,
		},
		{
			name: "INVALID - JSON marshal error",
			args: args{
				data: []InvalidPodInfo{
					{
						PODNAME:          "pod1",
						IPADDRESS:        "192.168.1.1",
						ENIID:            "eni-1234567890abcdef",
						SECURITYGROUPIDS: []string{"sg-12345678"},
						INVALIDFIELD:     func() {},
					},
				},
			},
			wantOut: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := OutputJSON(tt.args.data, out); (err != nil) != tt.wantErr {
				t.Errorf("OutputJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("OutputJSON() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
