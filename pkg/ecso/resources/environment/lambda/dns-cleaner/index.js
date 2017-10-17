const AWS = require('aws-sdk');

AWS.config.update({
    region: 'ap-southeast-2'
});

const ec2 = new AWS.EC2();
const r53 = new AWS.Route53();

/*
  Returns a promise for the DNS zone ID of the route 53 zone matching 'name'
*/
const getZoneId = name =>
    r53.listHostedZonesByName({
        DNSName: name
    }).promise().then(data => data.HostedZones[0].Id);

/*
  Filters array of record sets by suffix
  */
const filterRecordSets = (suffix, rs) =>
    rs.filter(r => {
        const i = r.Name.indexOf(suffix);
        return i > 0 && r.Name.length - suffix.length == i;
    });

/*
  Returns a promise that resolves to an array of route 53 resource record sets
  whose names conclude with the given suffix
  */
const getResourceRecordSets = (zoneId, suffix) => {
    const fetchPage = (records, startRecord) => {
        return r53.listResourceRecordSets({
            HostedZoneId: zoneId,
            StartRecordName: startRecord
        }).promise().then(data => {
            const rs = records.concat(data.ResourceRecordSets);
            return data.IsTruncated ? fetchPage(rs, data.NextRecordName) : rs;
        });
    };

    return fetchPage([]).then(rs => filterRecordSets(suffix, rs));
};

/*
  Returns a promise for an ec2 instance, or null if no instance with the requested
  private dns name exists
  */
const getInstanceByPrivateDnsName = name => {
    const params = {
        Filters: [{
            Name: "private-dns-name",
            Values: [name]
        }]
    };

    return ec2.describeInstances(params).promise().then(data => {
        if (data.Reservations.length) {
            return data.Reservations[0].Instances.length ? data.Reservations[0].Instances[0] : null;
        }
        return null;
    });
};

/*
  Returns a promise that resolves to the list of record sets rs filtered to only
  those which resolve to dead ec2 instances
  */
const findOrphanRecordSets = rs =>
    Promise.all(rs.map(r => isOrphanRecordSet(r).then(b => b ? r : null)))
    .then(rs => rs.filter(r => r != null));

/*
  Returns a promise for a boolean indicating if the provided recordset is orphaned
  An orphaned record set is one which resolved to either a non-existant ec2 instance,
  or an instance which is  no longer running
  */
const isOrphanRecordSet = recordSet =>
    Promise.all(recordSet.ResourceRecords.map(rr => {
        return getInstanceByPrivateDnsName(hostNameFromSrv(rr.Value));
    })).then(is => {
        return is.length == 0 || is.every(i => i == null || i.State.Name !== "running");
    });

/*
  Extracts the hostname section of an SRV record
  */
const hostNameFromSrv = srv => {
    const parts = srv.split(" ");
    return parts[parts.length - 1];
};

/*
  Returns a promise that resolves once all recordsets are deleted
  */
const deleteRecordSets = zoneId => recordSets =>
    recordSets.length ?
    executeChangeBatch(createChangeBatch(zoneId, "DELETE", recordSets)) :
    Promise.resolve("Nothing to delete");

/*
  Returns a promise for the result of calling the changeResourceRecordSets route 53
  API method
*/
const executeChangeBatch = params => {
    console.log("Executing change batch ", JSON.stringify(params));
    return r53.changeResourceRecordSets(params).promise();
};

/*
  Creates an r53 change batch for a list of resourceRecordSets
  */
const createChangeBatch = (zoneId, action, resourceRecordSets) => ({
    ChangeBatch: {
        Changes: resourceRecordSets.map(rs => createChange(action, rs)),
        Comment: "Deleted by ecso dns cleaner lambda"
    },
    HostedZoneId: zoneId
});

/*
  Creates an r53 change for a single resourceRecordSet
  */
const createChange = (action, resourceRecordSet) => ({
    Action: action,
    ResourceRecordSet: resourceRecordSet
});

/*
  Logs a message and data to console with standard json format
  */
const log = (message, data) => {
    console.log(JSON.stringify({
        message: message,
        data: data
    }));
};
/*
  Creates a function that returns a promise that resolves once all orphaned
  records that end with suffix have been deleted
  */
const cleanupZone = suffix => zoneId =>
    getResourceRecordSets(zoneId, suffix)
    .then(rs => {
        log("Checking records", rs);
        return rs;
    })
    .then(findOrphanRecordSets)
    .then(rs => {
        log("Found orphaned records", rs);
        return rs;
    })
    .then(deleteRecordSets(zoneId));
/*
  Returns a promise which resolves after deleting all orphaned records
  which contain the suffix in a r53 zone identified by zone name
  */
const cleanup = (zoneName, suffix) =>
    getZoneId(zoneName).then(cleanupZone(suffix));

/*
  Ensures zonename has trailing .
 */
const normalizeZoneName = x => x[x.length-1] === "." ? x : x + ".";

/*
  Builds a normalized zone name from parts
  */
const buildNormalizedZoneName = (...parts) => normalizeZoneName("." + parts.join(".")); 

/*
  Lambda entry point
*/
exports.handler = function(event, context, cb) {
    const zoneName = process.env.DNS_ZONE;
    const suffix = buildNormalizedZoneName(process.env.CLUSTER_NAME, zoneName);

    console.log(`Cleaning records with suffix ${suffix} from zone ${zoneName}`);
    console.log(JSON.stringify(event));

    cleanup(zoneName, suffix)
        .then(val => {
            console.log(JSON.stringify(val));
            cb(null, val);
        })
        .catch(err => {
            cb(err);
        });
};
