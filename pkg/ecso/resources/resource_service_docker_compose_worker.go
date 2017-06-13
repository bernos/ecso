package resources

import "text/template"

var WorkerServiceDockerComposeFile = NewTextFile("docker-compose.yaml", workerComposeFileTemplate)

var workerComposeFileTemplate = template.Must(template.New("workerComposeFile").Parse(`
version: '2'

volumes:
  nginxdata: {}

services:
  worker:
    image: busybox:latest
    mem_limit: 10000000
    volumes:
      - nginxdata:/nginx
    command: sh -c "while true; do echo \"This is the {{.Service.Name}} service <p><pre>` + "`env`" + `</pre></p> \" > /nginx/index.html; sleep 3; done"
`))
