package gojenkins

import (
	"errors"
	"strconv"
)

type Folder struct {
	Raw     *folderResponse
	Jenkins *Jenkins
	Base    string
}

type folderResponse struct {
	Actions           []generalObj
	Description       string      `json:"description"`
	DisplayName       string      `json:"displayName"`
	DisplayNameOrNull interface{} `json:"displayNameOrNull"`
	HealthReport      []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconUrl       string `json:"iconUrl"`
		Score         int64  `json:"score"`
	} `json:"healthReport"`
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	Jobs        []*job  `json:"jobs"`
	PrimaryView *view   `json:"primaryView"`
	Views       []*view `json:"views"`
}

func (f *Folder) GetName() string {
	return f.Raw.Name
}

func (f *Folder) GetDescription() string {
	return f.Raw.Description
}

func (f *Folder) GetDetails() *folderResponse {
	return f.Raw
}

func (f *Folder) Poll() (int, error) {
	_, err := f.Jenkins.Requester.GetJSON(f.Base, f.Raw, nil)
	if err != nil {
		return 0, err
	}
	return f.Jenkins.Requester.LastResponse.StatusCode, nil
}

func (f *Folder) GetAllJob() []*job {
	f.Poll()
	return f.Raw.Jobs
}

func (f *Folder) GetJob(id string) (*Job, error) {
	job := Job{Jenkins: f.Jenkins, Raw: new(jobResponse), Base: f.Base + "/job/" + id}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

// Create a new job from config File inside the folder.
// Method takes XML string as first parameter, and  name.
func (f *Folder) CreateJob(config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	}
	job := Job{Jenkins: f.Jenkins, Raw: new(jobResponse)}
	resp, err := f.Jenkins.Requester.PostXML(f.Base+"/createItem", config, job.Raw, qr)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (f *Folder) CopyJob(copyFrom string, newName string) (*Job, error) {
	qr := map[string]string{"name": newName, "from": copyFrom, "mode": "copy"}
	resp, err := f.Jenkins.Requester.Post(f.Base+"/createItem", nil, nil, qr)
	if err != nil {
		return nil, err
	}
	job := Job{Jenkins: f.Jenkins, Raw: new(jobResponse), Base: f.Base + "/job/" + newName}
	if resp.StatusCode == 200 {
		return &job, nil
	}
	newJob, err := f.GetJob(newName)
	if err != nil {
		return nil, err
	}
	return newJob, nil
}

// Delete a job.
func (f *Folder) DeleteJob(name string) (bool, error) {
	deleted, err := f.Jenkins.DeleteJob(f.GetName() + "/job/" + name)
	if deleted {
		return deleted, nil
	}
	// If Jenkins sends error then we should check if the job deleted
	// before sending the response. Jenkins can send 404 after deleting
	// the job and before sending the html.
	job, err := f.GetJob(name)
	if job == nil {
		return true, nil
	}
	return false, err
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
func (f *Folder) BuildJob(name string, options ...interface{}) (bool, error) {
	job := Job{Jenkins: f.Jenkins, Raw: new(jobResponse), Base: "/job/" + f.GetName() + "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (f *Folder) GetBuild(jobName string, number int64) (*Build, error) {
	job, err := f.GetJob(jobName)
	if err != nil {
		return nil, err
	}
	build := Build{Jenkins: f.Jenkins, Job: job, Raw: new(buildResponse), Depth: 1, Base: "/job/" + f.GetName() + "/job/" + jobName + "/" + strconv.FormatInt(number, 10)}
	status, err := build.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &build, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (f *Folder) GetAllBuildIds(job string) ([]jobBuild, error) {
	jobObj, err := f.GetJob(job)
	if err != nil {
		return nil, err
	}
	return jobObj.GetAllBuildIds()
}
