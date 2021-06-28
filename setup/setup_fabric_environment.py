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
ec2Client = boto3.client('ec2')
secretsManagerClient = boto3.client("secretsmanager")
s3Client = boto3.client("s3")

#Constants
certificateBucket = "reinvent2019-amb-artifacts-us-east-1"

#The three variables below are global that holds the context that we will be using through out this file.
blockchainNetwork = None
networkDetails = None

blockchainMember = None
memberDetails = None

peer = None

#Generates a random string, which is used for generating Tokens.
def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


def runCommand(args = []):
    print("Executing the following Command:" + ' '.join(args))

    process = subprocess.Popen(args,
                           stdout=subprocess.PIPE,
                           universal_newlines=True)

    return_code = None
    while True:
        output = process.stdout.readline()
        print(output.strip())
        # Do something else
        return_code = process.poll()
        if return_code is not None:
            # Process has finished, read rest of the output
            for output in process.stdout.readlines():
                print(output.strip())
            break
    return return_code

#============================= Network Selection ================================
blockchainNetworks = ambClient.list_networks(Status='AVAILABLE')

if len(blockchainNetworks) == 0:
    print("You currently are not a member of any Amazon Managed Blockchain Networks.")
    exit -1

print("--------------------------------------------------------")
print("Amazon Managed Blockchain Network Selection: ")
print("----------------------------------------------------------")
if len(blockchainNetworks['Networks']) == 0:
    print("You are not part of a Blockchain Network. Please join one.")
    sys.exit()
elif len(blockchainNetworks['Networks']) == 1:
    print("You currently are a member of only one Amazon Managed Blockchain Network.")
    print("Would you like to use Network: " + blockchainNetworks['Networks'][0]['Name'] + " [y/n]")
    tryAgain = True
    while tryAgain is True:
        choice = input()
        if choice == 'y' or choice == 'yes':
            blockchainNetwork = blockchainNetworks['Networks'][0]
            tryAgain = False
        elif choice == 'n' or choice == 'no':
            tryAgain = False
            print("Exiting.")
        else:
            print("Unexpected input. Please try again!")
else:
    count = 0
    for network in blockchainNetworks['Networks']:
        count = count + 1
        print (str(count) + ") " + network['Name'] + " (Id: " + network['Id'] + ")")

    print ("\nSelect the network that you want to setup this environment for: ")
    tryAgain = True
    while tryAgain is True:
        try:
            choice = input()
            blockchainNetwork = blockchainNetworks['Networks'][int(choice) - 1]
            tryAgain = False
        except KeyboardInterrupt:
            sys.exit()
        except Exception as e:
            print(e)
            print ("Invalid Input. Please try again. ")
    print ("\n\n You have chosen Network: " + blockchainNetwork['Name'])

networkDetails = ambClient.get_network(NetworkId=blockchainNetwork['Id'])

#============================= Member Selection ================================
blockchainMembers = ambClient.list_members(Status='AVAILABLE', NetworkId=blockchainNetwork['Id'], IsOwned=True)

print("--------------------------------------------------------")
print("Amazon Managed Blockchain Member Selection: ")
print("--------------------------------------------------------")
if len(blockchainNetworks['Networks']) == 0:
    print("You are not part of a Blockchain Network. Please join one.")
    sys.exit()
elif len(blockchainMembers['Members']) == 1:
    print("This account only has one member to chose from. Would you like to use Member: " + blockchainMembers['Members'][0]['Name'] + " [y/n]")
    tryAgain = True
    while tryAgain:
        choice = input()
        if choice == 'y' or choice == 'yes':
            blockchainMember = blockchainMembers['Members'][0]
            tryAgain = False
        elif choice == 'n' or choice == 'no':
            tryAgain = False
            print("Exiting.")
        else:
            print("Unexpected input. Please try again!")
else:
    count = 0
    for member in blockchainMembers['Members']:
        count = count + 1
        print (str(count) + ") " + member['Name'] + " (Id=" + member['Id'] + ")")

    print ("\nSelect the member that you want to setup this environment for: ")
    tryAgain = True
    while tryAgain:
        try:
            choice = input()
            blockchainMember = blockchainMembers['Members'][int(choice) - 1]
            tryAgain = False
        except KeyboardInterrupt:
            sys.exit()
        except Exception as e:
            print(e)
            print ("Invalid Input. Please try again. ")
    print ("\n\n You have chosen Member: " + blockchainMember['Name'])

memberDetails = ambClient.get_member(NetworkId=blockchainNetwork['Id'], MemberId=blockchainMember['Id'])

print ("You have chosen Network: " + blockchainNetwork['Name'] + " and Member: " + blockchainMember['Name'])

#============================= Peer Selection ================================
#Check if there are any peers, if not, we will ask to create it.
peers = ambClient.list_nodes(Status='AVAILABLE', NetworkId=blockchainNetwork['Id'], MemberId=blockchainMember['Id'])

print("--------------------------------------------------------")
print("Amazon Managed Blockchain Peer Selection: ")
print("--------------------------------------------------------")
if len(peers['Nodes']) == 0:
    print ("You do not have any peers in the select network and membership. Please create one using the following command and wait till its up before rerunnign this script.")
    print ("aws managedblockchain create-node --network-id " + blockchainNetwork['Id'] + " --member-id " + blockchainMember['Id'] + " --node-configuration 'InstanceType=bc.t3.medium,AvailabilityZone=us-east-1a'")
    sys.exit()
elif len(peers['Nodes']) == 1:
    print("This account only has one peer to chose from. Would you like to use Peer: " + peers['Nodes'][0]['Id'] + " [y/n]")
    tryAgain = True
    while tryAgain:
        choice = input()
        if choice == 'y' or choice == 'yes':
            peer = peers['Nodes'][0]
            tryAgain = False
        elif choice == 'n' or choice == 'no':
            tryAgain = False
            print("Exiting.")
        else:
            print("Unexpected input. Please try again!")
else:
    count = 0
    for peer in peers['Nodes']:
        count = count + 1
        print (str(count) + ") " + peer['Id'])

    print ("\nSelect the member that you want to setup this environment for: ")
    tryAgain = True
    while tryAgain:
        try:
            choice = input()
            peer = peer['Nodes'][int(choice) - 1]
            tryAgain = False
        except KeyboardInterrupt:
            sys.exit()
        except Exception as e:
            print(e)
            print ("Invalid Input. Please try again. ")
    print ("\n\n You have chosen Peer: " + peer['Name'])

peerDetails = ambClient.get_node(NetworkId=blockchainNetwork['Id'], MemberId=blockchainMember['Id'], NodeId=peer['Id'])

print ("We are going to execute a few steps in this script. We are going to: ")
print ("1) Setup a VPC Endpoint so that this machine can talk to your Blockchain Network.")
print ("2) Create an exports file that contains environment variables that will be useful to use so you don't have to remember them.")
print ("3) Obtain an admin Certificate. You will need to provide a password.")

input("Press Enter to continue...")

def waitTillVPCEndpointIsReady(vpcEndpointName, vpcId):
    keepChecking=True
    endPointStatus=None
    while (keepChecking):
        endPointStatus = ec2Client.describe_vpc_endpoints(Filters=[{'Name': 'service-name', 'Values': [vpcEndpointName]},{'Name': 'vpc-id', 'Values' :[vpcId]}])
        if endPointStatus['VpcEndpoints'][0]['State'] != 'pending':
            keepChecking=False
            break
        print("VPC Endpoint state is : " + endPointStatus['VpcEndpoints'][0]['State'] + ". Waiting. ")
        time.sleep(3)
    if endPointStatus['VpcEndpoints'][0]['State'] != 'available':
        print("VPC Endpoint Creation failed. State = " + endPointStatus['VpcEndpoints'][0]['State'])
        sys.exit()

def update_vpc_security_group(security_group_id):
    try:
        ec2Client.authorize_security_group_ingress(
            GroupId=security_group_id,
            IpPermissions=[
                {
                  'IpProtocol':'tcp',
                  'FromPort':30000,
                  'ToPort':31000,
                  'UserIdGroupPairs':[{'GroupId':security_group_id}]
                }
            ])
    except Exception as e:
        if 'InvalidPermission.Duplicate' in str(e):
            print("Duplicate SG Rule. Ignoring.")
        else:
            print("Attempt to add SG rule did not succeed.")
            print(e)


def setup_vpc_endpoint():
    print("--------------------------------------------------------")
    print ("Checking VPC Endpoints. ")
    print("--------------------------------------------------------")

    vpcEndpointName=networkDetails['Network']['VpcEndpointServiceName']

    response = requests.get('http://169.254.169.254/latest/meta-data/instance-id')
    instance_id = response.text

    instances = ec2Client.describe_instances(InstanceIds=[instance_id])
    if len(instances['Reservations'][0]['Instances']) != 1:
        print ("Unable to find instance id when calling EC2! " + instance_id)
    #There should only be one instance since we specified the instance id of this instance.
    thisEc2Instance = instances['Reservations'][0]['Instances'][0]

    #Check to make sure it doesn't already exist
    endpoints = ec2Client.describe_vpc_endpoints(Filters=[{'Name': 'service-name', 'Values': [vpcEndpointName]},{'Name': 'vpc-id', 'Values' :[thisEc2Instance['VpcId']]}])

    if len(endpoints['VpcEndpoints']) == 1:
        print("VPC Endpoint already exists.")
        subnetsToAdd=[]
        securityGroupsToAdd=[]
        if thisEc2Instance['SubnetId'] not in endpoints['VpcEndpoints'][0]['SubnetIds']:
            print ("VPC Endpoint does not include this machines subnet. Adding it now.")
            subnetsToAdd=[thisEc2Instance['SubnetId']]

        hasSgs=False
        for endPointSg in endpoints['VpcEndpoints'][0]['Groups']:
            if endPointSg in thisEc2Instance['SecurityGroups']:
                hasSgs = True
                break

        if hasSgs == False:
            print("VPC Endpoint doesn't have a shared SG. Adding now. ")
            securityGroupsToAdd=[thisEc2Instance['SecurityGroups'][0]['GroupId']]

        if len(subnetsToAdd) != 0 or len(securityGroupsToAdd) != 0:
            ec2Client.modify_vpc_endpoint(VpcEndpointId=endpoints['VpcEndpoints'][0]['VpcEndpointId'],
                                          AddSubnetIds=subnetsToAdd,
                                          AddSecurityGroupIds=securityGroupsToAdd)

        update_vpc_security_group(thisEc2Instance['SecurityGroups'][0]['GroupId'])

        waitTillVPCEndpointIsReady(vpcEndpointName, thisEc2Instance['VpcId'])
        print("--------------------------------------------------------")
    else:
        vpcCreateResponse = ec2Client.create_vpc_endpoint(
            VpcEndpointType='Interface',
            VpcId=thisEc2Instance['VpcId'],
            ServiceName=vpcEndpointName,
            SubnetIds=[thisEc2Instance['SubnetId']],
            SecurityGroupIds=[thisEc2Instance['SecurityGroups'][0]['GroupId']],
            ClientToken=randomString(),
            PrivateDnsEnabled=True
        )

        update_vpc_security_group(thisEc2Instance['SecurityGroups'][0]['GroupId'])

        waitTillVPCEndpointIsReady(vpcEndpointName, thisEc2Instance['VpcId'])

        print("--------------------------------------------------------")
        print("Successfully created VPC Endpoint. VPC Endpoint is : " + vpcCreateResponse['VpcEndpoint']['VpcEndpointId'] + ". Waiting till its ready")
        print("--------------------------------------------------------")

def create_export_file():

    print("--------------------------------------------------------")
    print("Creating export file")
    print("--------------------------------------------------------")

    f = open(expanduser("~") + "/fabric_exports", "w")
    f.writelines("export CAFILE=/opt/home/managedblockchain-tls-chain.pem" + "\n")
    f.writelines("export NETWORKID=" + networkDetails['Network']['Id'] + "\n")
    f.writelines("export MEMBERID=" + memberDetails['Member']['Id'] + "\n")
    f.writelines("export MSP=" + blockchainMember['Id'] + "\n")
    f.writelines("export MSP_PATH=/opt/home/admin-msp" + "\n")
    f.writelines("export ADMINUSER=" + memberDetails['Member']['FrameworkAttributes']['Fabric']['AdminUsername'] + "\n")
    f.writelines("export CASERVICEENDPOINT=" + memberDetails['Member']['FrameworkAttributes']['Fabric']['CaEndpoint'] + "\n")
    f.writelines("export ORDERER=" + networkDetails['Network']['FrameworkAttributes']['Fabric']['OrderingServiceEndpoint'] + "\n")
    f.writelines("export ORDERINGSERVICEENDPOINT=" + networkDetails['Network']['FrameworkAttributes']['Fabric']['OrderingServiceEndpoint'] + "\n")
    f.writelines("export PEERNODEID=" + peer['Id'] + "\n")
    f.writelines("export PEER=" + peerDetails['Node']['FrameworkAttributes']['Fabric']['PeerEndpoint'] + "\n")
    f.writelines("export PEERSERVICEENDPOINT=" + peerDetails['Node']['FrameworkAttributes']['Fabric']['PeerEndpoint'] + "\n")
    f.writelines("export PEEREVENTENDPOINT="+ peerDetails['Node']['FrameworkAttributes']['Fabric']['PeerEventEndpoint'] + "\n")
    f.writelines("export VPCENDPOINTSERVICENAME=" + networkDetails['Network']['VpcEndpointServiceName'] + "\n")
    f.close()

    print ("Generated ~/fabric_exports. Showing contents of the file:")
    with open(expanduser("~") + '/fabric_exports', 'r') as f:
        print(f.read())

    print("--------------------------------------------------------")
    print("NOTE: run 'source ~/fabric_exports' every time you start a new session. This will not be included in your bash profile.")
    input("Please press enter to continue.")

def obtain_admin_cert():
    print("Creating Admin Certificate")
    print("--------------------------------------------------------")

    print("Please enter your Admin password. Default password is 'Admin123'. :")
    adminPass = input()

    if len(adminPass) == 0:
        adminPass = "Admin123"

    binaryLocation="fabric-ca-client"
    # Check if we can find the hardcoded path, if so, use it. Otherwise, let the os find it in $PATH
    if os.path.exists("/home/ec2-user/go/src/github.com/hyperledger/fabric-ca/bin/fabric-ca-client") == True:
        binaryLocation="/home/ec2-user/go/src/github.com/hyperledger/fabric-ca/bin/fabric-ca-client"

    command = [binaryLocation, "enroll"
                , "-u", "https://" + memberDetails['Member']['FrameworkAttributes']['Fabric']['AdminUsername'] + ":" + adminPass + "@" + memberDetails['Member']['FrameworkAttributes']['Fabric']['CaEndpoint']
                , "--tls.certfiles", "/home/ec2-user/managedblockchain-tls-chain.pem"
                , "-M", "/home/ec2-user/admin-msp"]

    try:
        print ("Executing: " + ' '.join(command))
        subprocess.check_output(command, stderr=subprocess.STDOUT)
    except subprocess.CalledProcessError as e:
        print ("Failed to generate Admin Certificate! Please check output.")
        print (e)
        sys.exit()

    #Delete admin certs dir because we are going to create it.
    shutil.rmtree(expanduser("~") + "/admin-msp/admincerts", ignore_errors=True)
    #Now we have to fix up the certs a bit
    shutil.copytree(expanduser("~") + "/admin-msp/signcerts", expanduser("~") + "/admin-msp/admincerts")

    print("--------------------------------------------------------")
    print("Completed generating Admin Certificate. This certificate is used for admin operations such as channel creation, or creation of other identity certificates.")
    input("Press enter to continue.")

def store_public_certs():
    print("--------------------------------------------------------")
    print("Storing public certificates to S3 to share with other members")
    print("--------------------------------------------------------")

    adminMspPath = expanduser("~") + "/admin-msp/"
    adminPublicCerts3Key = blockchainNetwork['Id'] + "/" + blockchainMember['Id'] + "/admin-msp/admincerts/cert.pem"
    adminPublicCertFile = open(adminMspPath + "admincerts/cert.pem", "rb")

    #Copy Admin Public Cert to S3.
    s3Client.put_object(ACL='public-read', Bucket=certificateBucket,
                       Key=adminPublicCerts3Key,
                       Body=adminPublicCertFile)

    #Copy CA Public Certs to S3.
    caPublicCerts3Key = blockchainNetwork['Id'] + "/" + blockchainMember['Id'] + "/admin-msp/cacerts/ca-" + blockchainMember['Id'] + "-" + blockchainNetwork['Id'] + "-us-east-1-amazonaws-com.pem"
    caCertsPath = adminMspPath + "cacerts/"

    print ("Copying CA Certificates from " + caCertsPath + " to S3")
    for item in os.listdir(caCertsPath):
        fullItemPath = os.path.join(caCertsPath, item)
        print ("Found Item: " + item + " in " + caCertsPath)
        if os.path.isfile(fullItemPath):
            print ("Copying CA Certificate from: " + fullItemPath + " to: s3://" + certificateBucket + "/" + caPublicCerts3Key)
            caCertsFile = open(fullItemPath, "rb")
            s3Client.put_object(ACL='public-read',
                       Bucket=certificateBucket,
                       Key=caPublicCerts3Key,
                       Body=caCertsFile)
    input("Completed copying public certificates to S3. Please enter to continue... ")

def store_member_information():
    peerAddress = peerDetails['Node']['FrameworkAttributes']['Fabric']['PeerEndpoint']
    s3Client.put_object(ACL='public-read',
                        Body=peerAddress.encode(),
                        Bucket=certificateBucket,
                        Key=blockchainNetwork['Id'] + "/" + blockchainMember['Id'] + "/peer_address.txt"
                        )
    print("Copied Peer Address to S3 to: s3://" + certificateBucket + "/" + blockchainNetwork['Id'] + "/" + blockchainMember['Id'] + "/peer_address.txt")
    input("Press Enter to continue...")


setup_vpc_endpoint()
create_export_file()
obtain_admin_cert()
store_public_certs()
store_member_information()

print("All setup has been completed! Please run 'source ~/fabric_exports to pick up the environment variables. ")
