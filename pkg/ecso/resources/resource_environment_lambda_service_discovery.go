package resources

const ServiceDiscoveryLambdaVersion = "1.0.0"

var environmentServiceDiscoveryLambdaSource = `
processContainer = container => {
    (container.networkBindings || []).forEach(processNetworkBinding(container));
}

processNetworkBinding = container => binding => {
    if (container.lastStatus == "RUNNING") {
        ensureRecordExists(binding);
    } else if (container.lastStatus == "STOPPED") {
        ensureRecordDoesNotExist(binding);
    }
}

ensureRecordExists = binding => {
    console.log("Ensuring record exists for", binding);
}

ensureRecordDoesNotExist = binding => {
    console.log("Ensuring record does not exist for", binding);
}

exports.handler = function(event, context) {
    event.detail.containers.forEach(processContainer);
}
`
