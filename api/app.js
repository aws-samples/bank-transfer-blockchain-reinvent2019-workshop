/*
# Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
const connection = require('./connection.js');
const query = require('./query.js');
const invoke = require('./invoke.js');
const WebSocketServer = require('ws');
const express = require('express');
const bodyParser = require('body-parser');
const http = require('http');
const util = require('util');
const app = express();
const cors = require('cors');
const hfc = require('fabric-client');
const uuidv4 = require('uuid/v4');
const log4js = require('log4js');
const username = 'admin';

log4js.configure({
	appenders: {
		out: { type: 'stdout' },
	},
	categories: {
		default: { appenders: ['out'], level: 'info' },
	}
});

var logger = log4js.getLogger('BANKAPI');
hfc.addConfigFile('config.json');
const config = require('./config.json');
const host = 'localhost';
const port = 8081;
const orgName = config.org

var channelName = hfc.getConfigSetting('channelName');
var chaincodeName = hfc.getConfigSetting('chaincodeName');
var peers = hfc.getConfigSetting('peers');

///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// SET CONFIGURATIONS ///////////////////////////
///////////////////////////////////////////////////////////////////////////////
app.options('*', cors());
app.use(cors());
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({
	extended: false
}));
app.use(function(req, res, next) {
	logger.info(' ##### New request for URL %s', req.originalUrl);
	return next();
});

//wrapper to handle errors thrown by async functions. We can catch all
//errors thrown by async functions in a single place, here in this function,
//rather than having a try-catch in every function below. The 'next' statement
//used here will invoke the error handler function - see the end of this script
const awaitHandler = (fn) => {
	return async(req, res, next) => {
		try {
			await fn(req, res, next)
		}
		catch (err) {
			next(err)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// START SERVER /////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
var server = http.createServer(app).listen(port, function() {});
logger.info('****************** SERVER STARTED ************************');
logger.info('***************  Listening on: http://%s:%s  ******************', host, port);
server.timeout = 240000;

function getErrorMessage(field) {
	var response = {
		success: false,
		message: field + ' field is missing or Invalid in the request'
	};
	return response;
}

app.get('/health', awaitHandler(async(req, res) => {
	res.sendStatus(200);
}));


// account return the details of the account, it invokes the queryAccount chaincode function
app.get('/account/:accNumber', awaitHandler(async(req, res) => {
	let args = req.params;
	let fcn = "queryAccount";
	//username = req.params["accNumber"];

	let response = await connection.getRegisteredUser(username, orgName, true);

	logger.info('##### GET account details - username : ' + username);
	logger.info('##### GET account details - userOrg : ' + orgName);
	logger.info('##### GET account details - channelName : ' + channelName);
	logger.info('##### GET account details - chaincodeName : ' + chaincodeName);
	logger.info('##### GET account details - fcn : ' + fcn);
	logger.info('##### GET account details - args : ' + JSON.stringify(args));
	logger.info('##### GET account details - peers : ' + peers);

	res.header("Access-Control-Allow-Origin", "*");
	let message = await query.queryChaincode(peers, channelName, chaincodeName, [req.params["accNumber"]], fcn, username, orgName);
	res.send(message[0]);
}));


// transactions returns the history of an account, it invokes the getTransactionHistory chaincode function 
app.get('/transactions/:accNumber', awaitHandler(async(req, res) => {
	let args = req.params;
	let fcn = "getTransactionHistory";
	//username = req.params["accNumber"];

	let response = await connection.getRegisteredUser(username, orgName, true);

	logger.info('##### GET account details - username : ' + username);
	logger.info('##### GET account details - userOrg : ' + orgName);
	logger.info('##### GET account details - channelName : ' + channelName);
	logger.info('##### GET account details - chaincodeName : ' + chaincodeName);
	logger.info('##### GET account details - fcn : ' + fcn);
	logger.info('##### GET account details - args : ' + JSON.stringify(args));
	logger.info('##### GET account details - peers : ' + peers);

	let message = await query.queryChaincode(peers, channelName, chaincodeName, [req.params["accNumber"]], fcn, username, orgName);
	res.send(Object.values(message[0].History));
}));

// the transfer method invokes the transfer chaincode function to perform an intra or interbank transfer
app.post('/transfer', awaitHandler(async(req, res) => {
	var args = req.body;
	var fcn = "transfer";

	logger.info('================ POST on transfer');
	logger.info('##### POST for transfer - username : ' + username);
	logger.info('##### POST for transfer - userOrg : ' + orgName);
	logger.info('##### POST for transfer - channelName : ' + channelName);
	logger.info('##### POST for transfer - chaincodeName : ' + chaincodeName);
	logger.info('##### POST for transfer - fcn : ' + fcn);
	logger.info('##### POST for transfer - args : ' + JSON.stringify(args))
	logger.info('##### POST for transfer - peers : ' + peers);

	var array_args = []
	array_args[0] = args['FromAccNumber']
	array_args[1] = args['ToBankID']
	array_args[2] = args['ToAccNumber']
	array_args[3] = String(args['Amount'])

	let message = await invoke.invokeChaincode(peers, channelName, chaincodeName, array_args, fcn, username, orgName);
	res.send(message);
}));



app.use(function(error, req, res, next) {
	res.status(500).json({ error: error.toString() });
});
