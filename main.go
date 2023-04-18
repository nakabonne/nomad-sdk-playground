package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/nomad/api"
)

const jobID = "foo-job"

func main() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to new client: %v", err)
	}
	vars := &api.Variable{
		Namespace: "default",
		Path:      fmt.Sprintf("nomad/jobs/%s", jobID),
		Items: map[string]string{
			"mypassword": "yay",
		},
	}
	_, _, err = client.Variables().Create(vars, nil)
	if err != nil {
		log.Fatalf("failed to create var: %v", err)
	}

	job := buildJob(jobID)
	_, _, err = client.Jobs().Register(job, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func buildJob(id string) *api.Job {
	job := api.NewBatchJob(id, id, "global", 50)
	job.Datacenters = []string{"dc1"}

	group := buildTaskGroup()
	job.AddTaskGroup(group)
	return job
}

func buildTaskGroup() *api.TaskGroup {
	group := api.NewTaskGroup("tg1", 1)

	task := api.NewTask("task1", "raw_exec")
	task.Templates = append(task.Templates, &api.Template{
		EmbeddedTmpl: ptr(`
{{- with nomadVar "nomad/jobs/${NOMAD_JOB_ID}" -}}
echo My password is {{.mypassword}}
{{- end -}}
		`),
		DestPath:   ptr("foo.sh"),
		ChangeMode: ptr("noop"),
		Perms:      ptr("744"),
	})

	task.Config = map[string]interface{}{
		"command": "/bin/bash",
		"args": []interface{}{
			"-c",
			`./foo.sh`},
	}

	group.AddTask(task)
	return group
}

func ptr[T any](v T) *T {
	return &v
}
