# DB - database interface

Provides a high level pluggable abstraction for databases.

**This might be removed**

## Interface

Initial thoughts lie around a CRUD interface. The number of times 
one has to write CRUD on top of database libraries, having to think 
through schema and data modelling based on different databases is a 
pain. Going lower level than this doesn't pose any value.

Event sourcing can be tackled in a separate package.

## Supported Databases

- Cassandra
- MariaDB
- Elasticsearch
