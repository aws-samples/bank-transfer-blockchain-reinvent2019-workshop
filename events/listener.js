/*
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License").
# You may not use this file except in compliance with the License.
# A copy of the License is located at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# or in the "license" file accompanying this file. This file is distributed 
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either 
# express or implied. See the License for the specific language governing 
# permissions and limitations under the License.
#
*/

'use strict';

const kinesis_stream = 'bank-transfer-events'

const log4js = require('log4js');
const util = require('util');
const hfc = require('fabric-client');
const aws = require('aws-sdk');
const config = require('./config.json');


log4js.configure({
    appenders: {
        out: { type: 'stdout' },
    },
    categories: {
        default: { appenders: ['out'], level: 'info' },
    }
});

const kinesis = new aws.Kinesis({
    apiVersion: '2013-12-02',
    region: 'us-east-1'
});

var logger = log4js.getLogger('CHAINCODE-EVENT-LISTENER');


hfc.addConfigFile('config.json');

var username = 'admin';
var orgName = config.org
var eventName = "transfer-event";
var channelName = hfc.getConfigSetting('channelName');
var chaincodeName = hfc.getConfigSetting('chaincodeName');
var peers = hfc.getConfigSetting('peers');
const connection = require('./connection.js');


async function main() {
    logger.info('============ Starting Listening for Events');

    await connection.getRegisteredUser(username, orgName, true);
    const fabric_client = await connection.getClientForOrg(orgName, username);
    const channel = fabric_client.getChannel()
    const eventHub = channel.getChannelEventHubsForOrg()[0];
    eventHub.connect(true);
    logger.info('Listening for %s on %s using org %s', eventName, channel, orgName);

    eventHub.registerChaincodeEvent(chaincodeName, eventName,
        (event, block_num, txnid, status) => {
            console.log(event);

            var record = JSON.parse(event['payload'])
            console.log(record)

            var params = {
                Data: JSON.stringify(record) + "\n",
                PartitionKey: record['ToBankID'],
                StreamName: kinesis_stream
            };

            kinesis.putRecord(params, function(err, data) {
                if (err) console.log(err, err.stack); // an error occurred
                else console.log(data); // successful response
            });


        },
        (error) => {
            console.log('Failed to receive the chaincode event ::' + error);
        }
    );
}

main();
