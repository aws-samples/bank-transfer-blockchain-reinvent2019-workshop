#!/usr/bin/python
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

import boto3;
import sys, re, optparse
import copy, datetime, time
import requests
import random
import string
import os
from os.path import expanduser
import subprocess
import shutil

assert sys.version_info > (3,0)

ambClient = boto3.client('managedblockchain')
s3Client = boto3.client("s3")

#Constants
certificateBucket = "reinvent2019-amb-artifacts-us-east-1"

#get all members from the network.
networkid = os.getenv("NETWORKID", "")
memberid = os.getenv("MEMBERID", "")

if len(networkid) == 0 or len(memberid) == 0:
    print("NETWORKID or MEMBERID environment variables are not set! Please run 'source ~/fabric_exports'")
    sys.exit()

blockchainMembers = ambClient.list_members(Status='AVAILABLE', NetworkId=networkid, IsOwned=False)

membersFound=[]

print("We are using Blockchain Network Id: " + networkid + " and Member: " + memberid)
print("Please ensure that all other members in your network have run setup_fabric_environment before running this script! ")
input("Press enter to continue...")

for member in blockchainMembers['Members']:
    #No need to copy my own certificate
    if member['Id'] == memberid:
        continue
    if "Ignore" in member['Name'] or member['Name'] == "AmazonManagedBlockchainMember":
        print ("Ignoring Member: " + member['Name'] + " because it is not a part of this workshop.")
        continue

    #Get the Admin Public Cert for the member
    adminPublicCert = None
    try:
        adminPublicCerts3Key = networkid + "/" + member['Id'] + "/admin-msp/admincerts/cert.pem"
        adminPublicCert = s3Client.get_object(Bucket=certificateBucket, Key=adminPublicCerts3Key)
    except:
        print("Failed to find certificates for Member: " + member['Name'] + "(" + member['Id'] + ")")
        continue

    #Get the CA Public Cert for the member
    caPublicCert = None
    try:
        caPublicCerts3Key = networkid + "/" + member['Id'] + "/admin-msp/cacerts/ca-" + member['Id'] + "-" + networkid + "-us-east-1-amazonaws-com.pem"
        caPublicCert = s3Client.get_object(Bucket=certificateBucket, Key=caPublicCerts3Key)
    except:
        print("Failed to find certificates for Member: " + member['Name'] + "(" + member['Id'] + ")")
        continue

    #Delete any old directories.
    shutil.rmtree(expanduser("~") + "/" + member['Id'] + "-msp/", ignore_errors=True)

    #Create any directors.
    os.mkdir(expanduser("~") + "/" + member['Id'] + "-msp/")
    os.mkdir(expanduser("~") + "/" + member['Id'] + "-msp/admincerts")
    os.mkdir(expanduser("~") + "/" + member['Id'] + "-msp/cacerts")

    #Download the certificates
    s3Client.download_file(Bucket=certificateBucket, Key=adminPublicCerts3Key, Filename=expanduser("~") + "/" + member['Id'] + "-msp/admincerts/cert.pem")
    s3Client.download_file(Bucket=certificateBucket, Key=caPublicCerts3Key, Filename=expanduser("~") + "/" + member['Id'] + "-msp/cacerts/cacert.pem")
    membersFound.append(member['Name'] + ":" + member['Id'])

    print("Copying members peer address. ")
    s3Client.download_file(Bucket=certificateBucket,
                           Key=networkid + "/" + member['Id'] + "/peer_address.txt",
                           Filename=expanduser("~") + "/environment/peer-address-" + member['Id'] + ".txt")

if len(membersFound) == 0:
    print ("Did not find any members certificates! Please ensure that other participants have copied their member certificates from running setup_fabric_environment.")
else:
    print ("Found the following members: " + '\n'.join(membersFound))
    print ("Copied their certificates. If you are missing any certificates, ask other participants to setup their environment and run this script again.")

