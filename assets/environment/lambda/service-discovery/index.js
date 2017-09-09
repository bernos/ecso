const AWS = require('aws-sdk');

AWS.config.update({
    region: 'ap-southeast-2'
});

const ecs = new AWS.ECS();
const ec2 = new AWS.EC2();

const processContainer = (containerInstance, taskDefinition) => container => {
    switch (container.lastStatus) {
        case "RUNNING":
            return processContainerStarted(containerInstance, taskDefinition, container);
        case "STOPPED":
            return processContainerStopped(containerInstance, taskDefinition, container);
        default:
            return Promise.resolve();
    }
}

const processContainerStarted = (containerInstance, taskDefinition, container) => {
    const recs = resourceRecordSets(containerInstance, taskDefinition, container);
    return Promise.all(recs.map(ensureResourceRecordSetExists));
}

const processContainerStopped = (containerInstance, taskDefinition, container) => {
    const recs = resourceRecordSets(containerInstance, taskDefinition, container);
    return Promise.all(recs.map(ensureResourceRecordSetDoesNotExist));
}

const resourceRecordSets = (containerInstance, taskDefinition, container) => {
    const env = containerEnv(taskDefinition, container.name);

    return env.reduce((records, envVar) => {

        const info = envVarToServiceDiscoveryInfo(envVar);

        if (info != null) {
            const binding = findBinding(info.port, container.networkBindings);

            records.push({
                Name: info.name + "." + process.env.DNS_ZONE,
                Type: "SRV",
                TTL: 0,
                Weight: 1,
                ResourceRecords: [
                    "1 1 " + binding.containerPort + " " + containerInstance.PrivateDnsName
                ]
            });
        }

        return records;
    }, []);
}

const envVarToServiceDiscoveryInfo = envVar => {
    const parts = envVar.name.split("_");

    if (parts.length == 3 && parts[0] == "SERVICE" && parts[2] == "NAME") {
        return {
            name: envVar.name,
            port: parts[1]
        };
    }

    return null;
}

const containerEnv = (taskDefinition, name) =>
    taskDefinition.containerDefinitions.reduce((env, c) =>
        name == c.name ? c.environment : env, []);

const findBinding = (containerPort, bindings) => {
    return (bindings || []).reduce((binding, b) => {
        return String(b.containerPort) == String(containerPort) ? b : binding;
    }, {});
}

const ensureResourceRecordSetExists = binding => {
    console.log("Ensuring record exists for", binding);
    return new Promise((resolve, reject) => resolve(binding));
};

const ensureResourceRecordSetDoesNotExist = binding => {
    console.log("Ensuring record does not exist for", binding);
    return new Promise((resolve, reject) => resolve(binding));
};

const getTaskDefinition = arn => {
    return new Promise((resolve, reject) => {
        ecs.describeTaskDefinition({ taskDefinition: arn }, (err, data) => {
            err ? reject(err) : resolve(data);
        });
    });
};

const getEC2Instance = id => {
    const params = {
        InstanceIds: [id]
    }

    return new Promise((resolve, reject) => {
        ec2.describeInstances(params, (err, data) => {
            err ? reject(err) : resolve(data.Reservations[0].Instances.reduce((instance, i) => {
                return i.InstanceId == id ? i : instance;
            }, null));
        });
    })
}

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

const handleEvent = event => {
    return getContainerInstance("ecso-demo-dev", event.detail.containerInstanceArn)
        .then(instance => {
            return getTaskDefinition(event.detail.taskDefinitionArn)
                .then(data => {
                    return Promise.all(event.detail.containers.map(processContainer(instance, data.taskDefinition)));
                });
        })
}

exports.handler = function (event, context, cb) {
    handleEvent(event)
        .then(val => {
            cb(null, val);
        })
        .catch(err => {
            cb(err);
        });
};
