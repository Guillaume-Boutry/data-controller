# Data-controller

This project is responsible data management. The current supported database is Cassandra.

The enroller microservice sends the embeddings for insertion here.

The authenticator microservice gets the embeddings for authentication here.

## Build

A docker file is present to build it !

Just run:

```bash
docker build . -t registry.zouzland.com/data-controller:snapshot
```

## Configuration

The different configuration are the following:
```bash
CASSANDRA=cassandra-face-authent.cassandra # Url to cassandra service
KEYSPACE=Face # Keyspace
TABLE=tbl_face # Column family
```

Override them with environment variables !