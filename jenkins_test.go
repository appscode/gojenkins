package gojenkins

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	jenkins *Jenkins
)

func TestInit(t *testing.T) {
	jenkins = CreateJenkins("http://localhost:8080", "admin", "admin")
	_, err := jenkins.Init()
	assert.Nil(t, err, "Jenkins Initialization should not fail")
}

func TestCreateJobs(t *testing.T) {
	job1ID := "Folder2_test"
	job_data := getFileAsString("folder.xml")

	job1, err := jenkins.CreateJob(job_data, job1ID)
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, job1ID, job1.GetName())
}

func getFileAsString(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
