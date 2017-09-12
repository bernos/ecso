const AWS = require('aws-sdk');

AWS.config.update({
    region: 'ap-southeast-2'
});

const ecs = new AWS.ECS();
const ec2 = new AWS.EC2();
const r53 = new AWS.Route53();

const processContainer = (zoneName, containerInstance, taskDefinition) => container => {
    const createRecords = updateDnsForContainer("CREATE");
    const deleteRecords = updateDnsForContainer("DELETE");

    switch (container.lastStatus) {
        case "RUNNING":
            return createRecords(zoneName, containerInstance, taskDefinition, container);
        case "STOPPED":
            return deleteRecords(zoneName, containerInstance, taskDefinition, container);
        default:
            return Promise.resolve([]);
    }
}

const updateDnsForContainer = action => (zoneName, containerInstance, taskDefinition, container) =>
    getDnsZoneId(zoneName)
    .then(zoneId => Promise.all(
        containerResourceRecordSets(zoneName, containerInstance, taskDefinition, container)
        .map(createChangeBatch(action, zoneId))
        .map(changeResourceRecordSet)));

const createChangeBatch = (action, zoneId) => resourceRecordSet => ({
    ChangeBatch: {
        Changes: [{
            Action: action,
            ResourceRecordSet: resourceRecordSet
        }],
        Comment: "Created by ecso service discovery lambda"
    },
    HostedZoneId: zoneId
});

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

const containerResourceRecordSets = (zoneName, containerInstance, taskDefinition, container) => {
    const env = containerEnv(taskDefinition, container.name);

    return env.reduce((records, envVar) => {

        const info = envVarToServiceDiscoveryInfo(envVar);

        if (info != null) {
            records.push(
                containerResourceRecordSet(
                    info.name,
                    zoneName,
                    findBinding(info.port, container.networkBindings),
                    containerInstance));
        }

        return records;
    }, []);
}

const containerResourceRecordSet = (serviceName, dnsZone, networkBinding, containerInstance) => ({
    Name: serviceName + "." + dnsZone,
    Type: "SRV",
    TTL: 0,
    ResourceRecords: [{
        Value: srvRecord(1, 1, networkBinding.hostPort, containerInstance.PrivateDnsName)
    }]
});

const srvRecord = (priority, weight, port, hostname) =>
    priority + " " + weight + " " + port + " " + hostname;

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

const containerEnv = (taskDefinition, name) =>
    taskDefinition.containerDefinitions.reduce((env, c) =>
        name == c.name ? c.environment : env, []);

const findBinding = (containerPort, bindings) =>
    (bindings || []).reduce((binding, b) =>
        String(b.containerPort) == String(containerPort) ? b : binding, {});

const changeResourceRecordSet = params =>
    new Promise((resolve, reject) => {
        r53.changeResourceRecordSets(params, (err, data) => {
            err ? reject(err) : resolve(data);
        });
    });

const getTaskDefinition = arn =>
    new Promise((resolve, reject) => {
        ecs.describeTaskDefinition({
            taskDefinition: arn
        }, (err, data) => {
            err ? reject(err) : resolve(data.taskDefinition);
        });
    });

const getEC2Instance = id =>
    new Promise((resolve, reject) => {
        ec2.describeInstances({
            InstanceIds: [id]
        }, (err, data) => {
            err ? reject(err) : resolve(findInstanceById(id, data.Reservations[0].Instances));
        });
    });

const findInstanceById = (id, instances) =>
    instances.reduce((instance, i) => i.InstanceId === id ? i : instance, null);

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

const handleEvent = (zoneName, clusterName, event) => {
    switch (event["detail-type"]) {
        case "ECS Task State Change":
            return handleTaskStateChangeEvent(zoneName, clusterName, event);
        default:
            return Promise.reject("Unknown event type " + event["detail-type"]);
    }
}

const handleTaskStateChangeEvent = (zoneName, clusterName, event) => {
    return getContainerInstance(clusterName, event.detail.containerInstanceArn)
        .then(instance => {
            return getTaskDefinition(event.detail.taskDefinitionArn)
                .then(taskDefinition => {
                    return Promise.all(event.detail.containers.map(processContainer(zoneName, instance, taskDefinition)));
                });
        });
}

exports.handler = function(event, context, cb) {
    handleEvent(process.env.DNS_ZONE, process.env.CLUSTER_NAME, event)
        .then(val => {
            cb(null, val);
        })
        .catch(err => {
            cb(err);
        });
};
