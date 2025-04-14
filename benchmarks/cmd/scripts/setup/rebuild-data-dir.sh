#!/bin/bash

rm -rf /data

mkdir -p /data/mongo1_data
mkdir -p /data/mongo2_data
mkdir -p /data/kvrocks_data
mkdir -p /data/cassandra_data
mkdir -p /data/cassandra_commitlog
mkdir -p /data/cassandra_saved_caches
mkdir -p /data/couchdb_data