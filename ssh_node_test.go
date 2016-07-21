package gojenkins

import (
	"fmt"
	"github.com/golang/glog"
	"testing"
)

var njenkins *Jenkins

func init() {
	var err error
	njenkins, err = CreateJenkins("https://jenkins.appscode.com", "meta.system-admin", "ohmy").Init()
	if err != nil {
		fmt.Println(err)
		glog.Fatal(err)
	}
}

func TestNodeCreate(t *testing.T) {
	req := &SSHNodeCreateOptions{
		Name:        "test-from-go-2",
		Description: "hello",
		LabelString: "test",
		RemoteFS:    "/var/lib/jenkins",
		Executors:   2,

		CredentialId: "ebecf58d-4351-4569-b91c-72f0934f1411",

		Host: "45.55.138.156",
		Port: "22",
	}

	n, err := njenkins.CreateSSHNode(req)
	fmt.Println("node info", n, err)
}
