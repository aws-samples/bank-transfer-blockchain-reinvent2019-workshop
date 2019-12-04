#!/bin/bash

TRANSFER_URL=http://$(curl http://169.254.169.254/latest/meta-data/public-ipv4):8081/transfer
FROM_ACCT=0000001
FROM_BANK=0001
TO_ACCT=101010
TO_BANK=0005
while [ 1 ]; do
  TRANSFER=$(($RANDOM % 10))

  echo "Transfer 0.0$TRANSFER from account $FROM_ACCT to $TO_ACCT at $TO_BANK"
  curl ${TRANSFER_URL} \
    --data-binary "{\"FromAccNumber\":\"$FROM_ACCT\",\"ToBankID\":\"$TO_BANK\",\"ToAccNumber\":\"$TO_ACCT\",\"Amount\":0.0$TRANSFER}" \
    -H 'Content-Type: application/json'
  echo
  sleep 1
done
