package gojenkins

// TESTS written for APPSCODE.
import (
	"testing"

	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"log"
)

const (
	configs = `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <actions />
  <description />
  <keepDependencies>false</keepDependencies>
  <properties>
	<com.thetigerworks.jenkins.plugins.jenkinsauthz.AuthorizationMatrixProperty>
      <permission>com.cloudbees.plugins.credentials.CredentialsProvider.Update:banana</permission>
      <permission>hudson.model.View.Delete:banana</permission>
      <permission>hudson.model.Item.Create:banana</permission>
      <permission>hudson.model.Run.Delete:banana</permission>
      <permission>hudson.model.Item.Workspace:banana</permission>
      <permission>com.cloudbees.plugins.credentials.CredentialsProvider.Delete:banana</permission>
      <permission>com.cloudbees.plugins.credentials.CredentialsProvider.ManageDomains:banana</permission>
      <permission>hudson.model.View.Configure:banana</permission>
      <permission>hudson.model.Item.Configure:banana</permission>
      <permission>hudson.model.View.Read:banana</permission>
      <permission>hudson.model.View.Create:banana</permission>
      <permission>hudson.model.Item.Cancel:banana</permission>
      <permission>hudson.model.Item.Delete:banana</permission>
      <permission>hudson.model.Item.Read:banana</permission>
      <permission>com.cloudbees.plugins.credentials.CredentialsProvider.View:banana</permission>
      <permission>com.cloudbees.plugins.credentials.CredentialsProvider.Create:banana</permission>
      <permission>hudson.model.Item.Build:banana</permission>
      <permission>hudson.scm.SCM.Tag:banana</permission>
      <permission>hudson.model.Item.Move:banana</permission>
      <permission>hudson.model.Item.Discover:banana</permission>
      <permission>hudson.model.Run.Update:banana</permission>
    </com.thetigerworks.jenkins.plugins.jenkinsauthz.AuthorizationMatrixProperty>
  </properties>
  <scm class="hudson.scm.NullSCM" />
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers />
  <concurrentBuild>false</concurrentBuild>
  <builders />
</project>`
)

// Test Codes for appscode. This will check if the library is compatible with matrix auth
// and project based matrix security plugin. We use real jenkins for this.
var (
	fJenkins *Jenkins
)

// Initializing real jenkins for test, this is hackish. We can also use the `jenkins`
// to test with the mock tester. but mock tester will not provide us the phabricator
// login and militancy checkups.
func init() {
	fJenkins, _ = CreateJenkins("https://jenkins.appscode.com", "tigerworks.system-admin", "ohmy").Init()
	fmt.Println("Running APPSCODE tests.")
}

func TestGetFolder(t *testing.T) {
	fmt.Println("[TEST] Testing Get Folder")
	f, err := fJenkins.GetFolder("banana")
	assert.Equal(t, nil, err)
	assert.Equal(t, "banana", f.GetName())
	assert.Equal(t, "", f.GetDescription())
	assert.Equal(t, f.Raw, f.GetDetails())

	f, err = fJenkins.GetFolder("no-folder-found")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "404", err.Error())
}

func TestGetAllFolderJobs(t *testing.T) {
	fmt.Println("[TEST] Testing Get All Folder Job")
	f, _ := fJenkins.GetFolder("banana")
	jobs := f.GetJobs()
	for _, job := range jobs {
		fmt.Println(job.Name, job.Color)
	}
	assert.NotNil(t, jobs)
}

func TestGetFolderJob(t *testing.T) {
	fmt.Println("[TEST] Testing Get Job")
	f, _ := fJenkins.GetFolder("banana")
	jobName := "job-apple"
	j, err := f.GetJob(jobName)
	assert.Equal(t, j.Raw.Name, jobName)
	assert.Nil(t, err)

	jobName = "job-no-job"
	j, err = f.GetJob(jobName)
	assert.NotNil(t, err)
	assert.Nil(t, j)
}

func TestCreateFolderJob(t *testing.T) {
	fmt.Println("[TEST] Creating job")
	f, _ := fJenkins.GetFolder("banana")
	jobName := "job-apple-created-from-test"
	_, err := f.CreateJob(configs, jobName)
	assert.Nil(t, err)

	job, err := f.GetJob(jobName)
	assert.Nil(t, err)
	assert.Equal(t, job.GetName(), jobName)
	assert.Equal(t, job.Raw.Name, jobName)

	f.DeleteJob(jobName)
}

func TestCopyFolderJob(t *testing.T) {
	fmt.Println("[TEST] Testing Copy Job")
	f, _ := fJenkins.GetFolder("banana")
	_, err := f.CopyJob("banana-job-two", "job-banana-copy")
	assert.Nil(t, err)

	a, err := f.GetJob("job-banana-copy")
	assert.Equal(t, a.GetName(), "job-banana-copy")
	b, err := f.GetJob("banana-job-two")
	assert.Equal(t, a.GetConfig(), b.GetConfig())

	f.DeleteJob("job-banana-copy")
}

func TestDeleteFolderJob(t *testing.T) {
	fmt.Println("[TEST] Testing Delete job")
	f, _ := fJenkins.GetFolder("banana")
	jobName := "job-delete-job"

	job, err := f.CreateJob(configs, jobName)
	assert.Nil(t, err)
	assert.NotNil(t, job)
	errorHandler(err)
	fmt.Println("Job created for delete with name", jobName)

	b, err := f.DeleteJob(jobName)
	fmt.Println(b, err)
	assert.Nil(t, err)
	assert.True(t, b)

	job, err = f.GetJob(jobName)
	assert.Equal(t, err.Error(), "404")
	assert.Nil(t, job)

}

func TestBuildJob(t *testing.T) {
	fmt.Println("[TEST] Test Building job")
	f, _ := fJenkins.GetFolder("banana")
	data, err := f.BuildJob("banana-job-two")
	fmt.Println(data, err)
	assert.True(t, data)
	assert.Nil(t, err)
}

func TestFolderBuildJob(t *testing.T) {
	fmt.Println("[TEST] Getting job Build")
	f, _ := fJenkins.GetFolder("banana")
	data, err := f.GetBuild("banana-job-two", 1)
	fmt.Println(data.Raw, err, data.GetConsoleOutput())
	assert.NotNil(t, data)
	assert.Nil(t, err)
}

func TestFolderAllBuild(t *testing.T) {
	fmt.Println("[TEST] Getting all job Build")
	f, _ := fJenkins.GetFolder("banana")
	ids, err := f.GetAllBuildIds("banana-job-two")
	assert.Nil(t, err)
	assert.NotNil(t, ids)
	for _, id := range ids {
		fmt.Println(id)
	}
}

func errorHandler(err error) {
	if err != nil {
		fmt.Println("Error Occured", err)
		os.Exit(1)
	}
}

func TestGetInnerJobs(t *testing.T) {
	fmt.Println("[TEST] Testing Inner Job")
	f, err := fJenkins.GetFolder("tigerworks")
	if err != nil {
		log.Fatal(err)
	}

	jobs := f.GetJobs()
	for _, job := range jobs {
		fmt.Println(job.Name)
	}
	fmt.Println("==================================")
	jobs = f.GetInnerJobs("/job/this-is-folder/")
	for _, job := range jobs {
		fmt.Println(job.Name)
	}
	fmt.Println("==================================")
	jobs = f.GetInnerJobs("")
	for _, job := range jobs {
		fmt.Println(job.Name)
	}
	fmt.Println("==================================")
	jobs = f.GetInnerJobs("/job/tamal/")
	for _, job := range jobs {
		fmt.Println(job.Name)
	}
}
