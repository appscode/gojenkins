// Copyright 2014 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Gojenkins is a Jenkins Client in Go, that exposes the jenkins REST api in a more developer friendly way.
package gojenkins

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Raw       *executorResponse
	Requester *Requester
}

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init Method. Should be called after creating a Jenkins Instance.
// e.g jenkins := CreateJenkins("url").Init()
// HTTP Client is set here, Connection to jenkins is tested here.
func (j *Jenkins) Init() (*Jenkins, error) {
	j.initLoggers()
	// Skip SSL Verification?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !j.Requester.SslVerify},
	}

	if j.Requester.Client == nil {
		cookies, _ := cookiejar.New(nil)

		client := &http.Client{
			Transport: tr,
			Jar:       cookies,
		}
		j.Requester.Client = client
	}

	// Check Connection
	j.Raw = new(executorResponse)
	_, err := j.Requester.GetJSON("/", j.Raw, nil)
	if err != nil {
		return nil, err
	}
	j.Version = j.Requester.LastResponse.Header.Get("X-Jenkins")
	if j.Raw == nil {
		panic("Connection Failed, Please verify that the host and credentials are correct.")
	}
	return j, nil
}

func (j *Jenkins) initLoggers() {
	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Get Basic Information About Jenkins
func (j *Jenkins) Info() *executorResponse {
	j.Requester.Do("GET", "/", nil, j.Raw, nil)
	return j.Raw
}

// Create a new Node
func (j *Jenkins) CreateNode(name string, numExecutors int, description string, remoteFS string, options ...interface{}) (*Node, error) {
	node := j.GetNode(name)
	if node != nil {
		return node, nil
	}
	node = &Node{Jenkins: j, Raw: new(nodeResponse), Base: "/computer/" + name}
	NODE_TYPE := "hudson.slaves.DumbSlave$DescriptorImpl"
	MODE := "NORMAL"
	qr := map[string]string{
		"name": name,
		"type": NODE_TYPE,
		"json": makeJson(map[string]interface{}{
			"name":               name,
			"nodeDescription":    description,
			"remoteFS":           remoteFS,
			"numExecutors":       numExecutors,
			"mode":               MODE,
			"type":               NODE_TYPE,
			"retentionsStrategy": map[string]string{"stapler-class": "hudson.slaves.RetentionStrategy$Always"},
			"nodeProperties":     map[string]string{"stapler-class-bag": "true"},
			"launcher":           map[string]string{"stapler-class": "hudson.slaves.JNLPLauncher"},
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
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

// Create a new job from config File
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
// e.g jenkins.CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	}
	job := Job{Jenkins: j, Raw: new(jobResponse)}
	job.Create(config, qr)
	return &job, nil
}

// Rename a job.
// First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (j *Jenkins) CopyJob(copyFrom string, newName string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + newName}
	return job.Copy(copyFrom, newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(name string) (bool, error) {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + name}
	return job.Delete()
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
func (j *Jenkins) BuildJob(name string, options ...interface{}) (bool, error) {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (j *Jenkins) GetNode(name string) *Node {
	node := Node{Jenkins: j, Raw: new(nodeResponse), Base: "/computer/" + name}
	if node.Poll() == 200 {
		return &node
	}
	return nil
}

func (j *Jenkins) GetBuild(jobName string, number int64) (*Build, error) {
	job, err := j.GetJob(jobName)
	if err != nil {
		return nil, err
	}
	return job.GetBuild(number)
}

func (j *Jenkins) GetJob(id string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + id}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

// Get the folder by name. Folder will be initiated by
// Cloudbees folder plugin.
func (j *Jenkins) GetFolder(id string) (*Folder, error) {
	folder := Folder{Jenkins: j, Raw: new(folderResponse), Base: "/job/" + id}
	status, err := folder.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &folder, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetAllNodes() []*Node {
	computers := new(Computers)
	j.Requester.GetJSON("/computer", computers, nil)
	nodes := make([]*Node, len(computers.Computers))
	for i, node := range computers.Computers {
		name := node.DisplayName
		// Special Case - Master Node
		if name == "master" {
			name = "(master)"
		}
		nodes[i] = j.GetNode(name)
	}
	return nodes
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)
func (j *Jenkins) GetAllBuildIds(job string) ([]jobBuild, error) {
	jobObj, err := j.GetJob(job)
	if err != nil {
		return nil, err
	}
	return jobObj.GetAllBuildIds()
}

// Get Only Array of Job Names, Color, URL
// Does not query each single Job.
func (j *Jenkins) GetAllJobNames() []job {
	exec := Executor{Raw: new(executorResponse), Jenkins: j}
	j.Requester.GetJSON("/", exec.Raw, nil)
	return exec.Raw.Jobs
}

// Get All Possible Job Objects.
// Each job will be queried.
func (j *Jenkins) GetAllJobs() ([]*Job, error) {
	exec := Executor{Raw: new(executorResponse), Jenkins: j}
	j.Requester.GetJSON("/", exec.Raw, nil)
	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		ji, err := j.GetJob(job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

// Returns a Queue
func (j *Jenkins) GetQueue() *Queue {
	q := &Queue{Jenkins: j, Raw: new(queueResponse), Base: j.GetQueueUrl()}
	q.Poll()
	return q
}

func (j *Jenkins) GetQueueUrl() string {
	return "/queue"
}

// Get Artifact data by Hash
func (j *Jenkins) GetArtifactData(id string) *fingerPrintResponse {
	fp := Fingerprint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(fingerPrintResponse)}
	return fp.GetInfo()
}

// Returns the list of all plugins installed on the Jenkins server.
// You can supply depth parameter, to limit how much data is returned.
func (j *Jenkins) GetPlugins(depth int) *Plugins {
	p := Plugins{Jenkins: j, Raw: new(pluginResponse), Base: "/pluginManager", Depth: depth}
	p.Poll()
	return &p
}

// Check if the plugin is installed on the server.
// Depth level 1 is used. If you need to go deeper, you can use GetPlugins, and iterate through them.
func (j *Jenkins) HasPlugin(name string) *Plugin {
	p := j.GetPlugins(1)
	return p.Contains(name)
}

// Verify Fingerprint
func (j *Jenkins) ValidateFingerPrint(id string) bool {
	fp := Fingerprint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(fingerPrintResponse)}
	if fp.Valid() {
		Info.Printf("Jenkins says %s is valid", id)
		return true
	}
	return false
}

func (j *Jenkins) GetView(name string) *View {
	url := "/view/" + name
	view := View{Jenkins: j, Raw: new(viewResponse), Base: url}
	view.Poll()
	return &view
}

func (j *Jenkins) GetAllViews() []*View {
	j.Poll()
	views := make([]*View, len(j.Raw.Views))
	for i, v := range j.Raw.Views {
		views[i] = j.GetView(v.Name)
	}
	return views
}

// Create View
// First Parameter - name of the View
// Second parameter - Type
// Possible Types:
// 		gojenkins.LIST_VIEW
// 		gojenkins.NESTED_VIEW
// 		gojenkins.MY_VIEW
// 		gojenkins.DASHBOARD_VIEW
// 		gojenkins.PIPELINE_VIEW
// Example: jenkins.CreateView("newView",gojenkins.LIST_VIEW)

func (j *Jenkins) CreateView(name string, viewType string) (bool, error) {
	exists := j.GetView(name)
	if exists != nil {
		Error.Println("View Already exists.")
		return false, errors.New("View already exists")
	}
	view := View{Jenkins: j, Raw: new(viewResponse), Base: "/view/" + name}
	url := "/createView"
	data := map[string]string{
		"name":   name,
		"type":   viewType,
		"Submit": "OK",
		"json": makeJson(map[string]string{
			"name": name,
			"mode": viewType,
		}),
	}
	r, err := j.Requester.Post(url, nil, view.Raw, data)
	if err != nil {
		return false, err
	}
	if r.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New(strconv.Itoa(r.StatusCode))
}

func (j *Jenkins) Poll() int {
	j.Requester.GetJSON("/", j.Raw, nil)
	return j.Requester.LastResponse.StatusCode
}

// Creates a new Jenkins Instance
// Optional parameters are: username, password
// After creating an instance call init method.
func CreateJenkins(base string, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: false, Headers: http.Header{}}
	if len(auth) == 2 {
		j.Requester.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}
