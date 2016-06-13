package gojenkins

import (
	"errors"
	"strconv"
	"net/url"
)

type SSHNodeCreateOptions struct {
	Name        string
	Description string
	Executors   int
	RemoteFS    string
	LabelString string
	Mode        string

	Host         string
	Port         string
	CredentialId string
}

func (j *Jenkins) CreateSSHNode(obj *SSHNodeCreateOptions) (*Node, error) {
	node := j.GetNode(obj.Name)
	if node != nil {
		return nil, errors.New("node already exists with this name")
	}
	node = &Node{Jenkins: j, Raw: new(nodeResponse), Base: "/computer/" + obj.Name}
	NODE_TYPE := "hudson.slaves.DumbSlave$DescriptorImpl"
	if obj.Mode == "" {
		obj.Mode = "NORMAL" // EXCLUSIVE
	}
	qr := map[string]string{
		"name": obj.Name,
		"type": NODE_TYPE,
		"json": makeJson(map[string]interface{}{
			"name":            obj.Name,
			"nodeDescription": obj.Description,
			"remoteFS":        obj.RemoteFS,
			"numExecutors":    obj.Executors,
			"labelString":     url.QueryEscape(obj.LabelString),
			"mode":            obj.Mode,
			"type":            NODE_TYPE,
			"": []string{
				"hudson.plugins.sshslaves.SSHLauncher",
				"hudson.slaves.RetentionStrategy$Always",
			},
			"retentionsStrategy": map[string]string{
				"stapler-class": "hudson.slaves.RetentionStrategy$Always",
				"$class":        "hudson.slaves.RetentionStrategy$Always",
			},
			"nodeProperties": map[string]string{"stapler-class-bag": "true"},
			"launcher": map[string]string{
				"stapler-class":        "hudson.plugins.sshslaves.SSHLauncher",
				"$class":               "hudson.plugins.sshslaves.SSHLauncher",
				"host":                 obj.Host,
				"credentialsId":        obj.CredentialId,
				"port":                 obj.Port,
				"launchTimeoutSeconds": "900",
				"maxNumRetries":        "30",
				"retryWaitTime":        "30",
			},
		}),
	}

	resp, err := j.Requester.GetXML("/computer/doCreateItem", nil, qr)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 400 {
		node.Poll()
		return node, nil
	}

	// try confirming if the nod presents in the node list
	node = j.GetNode(obj.Name)
	if node != nil {
		return node, nil
	}

	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}
