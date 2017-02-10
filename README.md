# ecso

# Installing

`go install github.com/bernos/ecso`

# Getting started

```bash
# Create a folder for your project
mkdir ~/my-project && cd ~/my-project

# Initialise a new ecso project
ecso init

# Set up a new ecso environment to deploy to
ecso environment add my-environment
ecso environment up my-environment

# Create a new ecso service and deploy it to the environment
ecso service add my-service
ecso service up my-service --environment my-environment

```
