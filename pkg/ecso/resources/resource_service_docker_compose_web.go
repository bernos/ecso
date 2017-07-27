package resources

var webServiceComposeFileTemplate = `
version: '2'

volumes:
  nginxdata: {}

services:
  web:
    image: nginx:latest
    mem_limit: 20000000
    ports:
      - "0:{{.Service.Port}}"
    volumes:
      - nginxdata:/usr/share/nginx/html/:ro
    command: /bin/bash -c "echo \"server { location / { root /usr/share/nginx/html; try_files \$$uri /index.html =404; } }\" > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"
  backend:
    image: busybox:latest
    mem_limit: 10000000
    volumes:
      - nginxdata:/nginx
    command: sh -c "while true; do echo \"This is the {{.Service.Name}} service <p><pre>` + "`env`" + `</pre></p> \" > /nginx/index.html; sleep 3; done"
`
