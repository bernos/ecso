const AWS = require('aws-sdk');

AWS.config.update({
    region: 'ap-southeast-2'
});

const ecs = new AWS.ECS();
const ec2 = new AWS.EC2();
const r53 = new AWS.Route53();

/*
  Creates a route 53 change batch
  */
const changeBatch = (action, zoneId, resourceRecordSet) => ({
    ChangeBatch: {
        Changes: [{
            Action: action,
            ResourceRecordSet: resourceRecordSet
        }],
        Comment: "Created by ecso service discovery lambda"
    },
    HostedZoneId: zoneId
});

/*
  Returns a promise for the DNS zone ID of the route 53 zone matching 'name'
  */
const getDnsZoneId = name => {
    var params = {
        DNSName: name
    };

    return new Promise((resolve, reject) => {
        r53.listHostedZonesByName(params, (err, data) => {
            err ? reject(err) : resolve(data.HostedZones[0].Id);
        });
    });
}

/*
  Creates an array of route 53 resource record sets for a given container. A record
  set is created for each port and service name indicated by the presence of
  SERVICE_<PORT>_NAME environment variables in the task definition
  */
const containerResourceRecordSets = (zoneName, containerInstance, taskDefinition, container) =>
    serviceDiscoveryInfo(container, taskDefinition)
    .map(info => containerResourceRecordSet(info.name, zoneName, findBinding(info.port, container.networkBindings), containerInstance, container.containerArn));

/*
  Returns an array of service discovery info object parsed from the env vars of
  a container. Each info object has the form { name: string, port: number }
  */
const serviceDiscoveryInfo = (container, taskDefinition) =>
    containerEnv(taskDefinition, container.name)
    .map(envVarToServiceDiscoveryInfo)
    .filter(info => info != null);

/*
  Creates a resource record set for a container and network binding
  */
const containerResourceRecordSet = (serviceName, dnsZone, networkBinding, containerInstance, setIdentifier) => ({
    Name: serviceName + "." + dnsZone,
    Type: "SRV",
    TTL: 0,
    SetIdentifier: setIdentifier,
    Weight: 1,
    ResourceRecords: [{
        Value: srvRecord(1, 1, networkBinding.hostPort, containerInstance.PrivateDnsName)
    }]
});

/*
  Creates an SRV DNS record
  */
const srvRecord = (priority, weight, port, hostname) =>
    priority + " " + weight + " " + port + " " + hostname;

/*
  Transforms an environment variable in the format SERVICE_<PORT>_NAME=<my.service>
  into an object in the form { name: my.service, port: <PORT> }. If the environment
  variable cannot be parsed, null is returned
  */
const envVarToServiceDiscoveryInfo = envVar => {
    const parts = envVar.name.split("_");

    if (parts.length == 3 && parts[0] == "SERVICE" && parts[2] == "NAME") {
        return {
            name: envVar.value,
            port: parts[1]
        };
    }

    return null;
}

/*
  Returns an array of environment variables for a given container in a
  task definition
  */
const containerEnv = (taskDefinition, name) =>
    taskDefinition.containerDefinitions.reduce((env, c) =>
        name == c.name ? c.environment : env, []);


/*
  Returns true if the the container has any service discovery env vars
  */
const hasServiceDiscoveryEnvVar = (container, taskDefinition) =>
    serviceDiscoveryInfo(container, taskDefinition).length > 0;

/*
  Retrives a binding by container port from a list of bindings
  */
const findBinding = (containerPort, bindings) =>
    (bindings || []).reduce((binding, b) =>
        String(b.containerPort) == String(containerPort) ? b : binding, {});

/*
  Returns a promise for the result of calling the changeResourceRecordSets route 53
  API method
  */
const executeChangeBatch = params => {
    console.log("Executing change batch ", JSON.stringify(params));
    return new Promise((resolve, reject) => {
        r53.changeResourceRecordSets(params, (err, data) => {
            err ? reject(err) : resolve(data);
        });
    });
}

/*
  Returns a promise for the task definition indicated by arn
  */
const getTaskDefinition = arn =>
    new Promise((resolve, reject) => {
        ecs.describeTaskDefinition({
            taskDefinition: arn
        }, (err, data) => {
            err ? reject(err) : resolve(data.taskDefinition);
        });
    });

/*
  Returns a promise for an ec2 instance matched by id
  */
const getEC2Instance = id =>
    new Promise((resolve, reject) => {
        ec2.describeInstances({
            InstanceIds: [id]
        }, (err, data) => {
            err ? reject(err) : resolve(findInstanceById(id, data.Reservations[0].Instances));
        });
    });

/*
  Finds an ec2 instance by id from an array of ec2 instances
  */
const findInstanceById = (id, instances) =>
    instances.reduce((instance, i) => i.InstanceId === id ? i : instance, null);

/*
  Returns a promise for an ec2 container instance with the given arn, who is a
  member of the given ecs cluster
  */
const getContainerInstance = (cluster, arn) => {
    const params = {
        cluster: cluster,
        containerInstances: [arn]
    }

    return new Promise((resolve, reject) => {
        ecs.describeContainerInstances(params, (err, data) => {
            if (err) {
                reject(err);
            } else if (!data.containerInstances.length) {
                resolve(null);
            } else {
                getEC2Instance(data.containerInstances[0].ec2InstanceId)
                    .then(resolve)
                    .catch(reject);
            }
        });
    });
}

/*
  Returns a promise for the result of handling a single ecs task change event
  */
const handleEvent = (zoneName, clusterArn, event) => {
    const desiredState = isTaskStartedEvent(event) ? "RUNNING" : "STOPPED";

    return Promise.all([
        getContainerInstance(clusterArn, event.detail.containerInstanceArn),
        getTaskDefinition(event.detail.taskDefinitionArn)
    ]).then(([instance, taskDefinition]) =>
        updateDnsForContainers(
            desiredState,
            filterDiscoverableContainers(event.detail.containers, taskDefinition),
            zoneName,
            instance,
            taskDefinition));
}

/*
  Returns a promise that will resolve after updating the DNS entries for a list
  of containers in a single task definition
  */
const updateDnsForContainers = (desiredState, containers, zoneName, instance, taskDefinition) =>
    Promise.all(containers.map(updateDnsForContainer(desiredState, zoneName, instance, taskDefinition)));

/*
  Returns a promise that will resolve after updating the DNS entries for a single
  container in a task definition
  */
const updateDnsForContainer = (desiredState, zoneName, containerInstance, taskDefinition) => container =>
    containerChangeBatches(desiredState, zoneName, containerInstance, taskDefinition, container)
    .then(executeChangeBatches);

/*
  Filters non-discoverable containers from an array of containers
  */
const filterDiscoverableContainers = (containers, taskDefinition) =>
    containers.filter(c => isDiscoverable(c, taskDefinition));

/*
  Returns true if a conainer is discoverable. A container must contain at least one network binding,
  at least one service discovery env var and be in a terminal state in order to be considered
  discoverable
  */
const isDiscoverable = (container, taskDefinition) =>
    container.networkBindings != null &&
    hasServiceDiscoveryEnvVar(container, taskDefinition) &&
    (container.lastStatus === "RUNNING" || container.lastStatus === "STOPPED");

/*
  Returns true if an event is actionable. Curently only task started and task stopped events are
  considered actionable
  */
const isActionableEvent = event =>
    isTaskStartedEvent(event) || isTaskStoppedEvent(event);

/*
  Returns true if both the desiredStatus and lastStatus of the task are RUNNING
  */
const isTaskStartedEvent = event =>
    isTaskStateChangedEvent(event) &&
    (event.detail.desiredStatus === event.detail.lastStatus) &&
    (event.detail.lastStatus === "RUNNING");

/*
  Returns true if the desiredStatus of the task is STOPPED
  */
const isTaskStoppedEvent = event =>
    isTaskStateChangedEvent(event) && (event.detail.desiredStatus === "STOPPED");

/*
  Returns true if the event is a task changed event
  */
const isTaskStateChangedEvent = event =>
    event["detail-type"] === "ECS Task State Change";

/*
  Returns a promise that wil resolve when all route 53 change batches have been
  executed
  */
const executeChangeBatches = changeBatches =>
    Promise.all(changeBatches.map(executeChangeBatch));

/*
  Returns a promise for an array of route 53 change batches for a container. The action for each
  change set will be determined by the desiredState and the last known container state
  */
const containerChangeBatches = (desiredState, zoneName, containerInstance, taskDefinition, container) =>
    getDnsZoneId(zoneName)
    .then(zoneId =>
        containerResourceRecordSets(zoneName, containerInstance, taskDefinition, container)
        .map(rs => changeBatch(containerAction(desiredState, container), zoneId, rs)));

/*
  Converts a desiredState and last known container state into an appropriate route53 changeset
  action. If either the desired or know state are STOPPED then return the delete action, otherwise
  UPSERT
  */
const containerAction = (desiredState, container) =>
    (container.lastStatus === "STOPPED" || desiredState === "STOPPED") ? "DELETE" : "UPSERT";

/*
  Lambda entry point
  */
exports.handler = function(event, context, cb) {
    console.log(JSON.stringify(event));

    if (isActionableEvent(event)) {
        handleEvent(process.env.DNS_ZONE, process.env.CLUSTER_ARN, event)
            .then(val => {
                console.log(JSON.stringify(val));
                cb(null, val);
            })
            .catch(err => {
                cb(err);
            });
    } else {
        cb(null, "Skipping unhandleable event");
    }
};
