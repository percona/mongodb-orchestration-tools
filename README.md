# dcos-mongo-tools

[![](https://godoc.org/github.com/percona/dcos-mongo-tools?status.svg)](http://godoc.org/github.com/percona/dcos-mongo-tools)
[![Build Status](https://travis-ci.org/percona/dcos-mongo-tools.svg?branch=master)](https://travis-ci.org/percona/dcos-mongo-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/percona/dcos-mongo-tools)](https://goreportcard.com/report/github.com/percona/dcos-mongo-tools)
[![codecov](https://codecov.io/gh/percona/dcos-mongo-tools/branch/master/graph/badge.svg)](https://codecov.io/gh/percona/dcos-mongo-tools)

Go-based tools for the [DC/OS 'percona-server-mongodb' service](https://docs.mesosphere.com/services/percona-server-mongodb/).

*Note: This code is intended for a specific integration/use case, therefore it is unlikely Issues or Pull Requests will be accepted from the public. Please fork if this is a concern.*

**Tools**:
- **mongodb-executor**: tool for executing tasks on the local mongod/mongos container
- **mongodb-controller**: tool for controlling the replica set initiation and adding system MongoDB users
- **mongodb-healthcheck**: tool for running DC/OS health and readiness checks on a MongoDB task
- **mongodb-watchdog**: daemon to monitor dcos pod status and manage mongodb replica set membership

## Use Case / Required
The tools in this repository are designed to be used specifically within the [DC/OS 'percona-server-mongodb' service](https://docs.mesosphere.com/services/percona-server-mongodb/) by using the [DC/OS SDK API](https://mesosphere.github.io/dcos-commons/reference/swagger-api/), etc.

The minimum requirements are:
1. DC/OS 1.10+
2. DC/OS SDK 0.42.1+

## Build
1. Install go1.10+ and 'make'
2. Run 'make' in git directory
3. Find binaries in 'bin' directory

## Contact
- Tim Vaillancourt - [Github](https://github.com/timvaillancourt) [Email](mailto:tim.vaillancourt@percona.com)
- Percona - [Twitter](https://twitter.com/Percona) [Contact Page](https://www.percona.com/about-percona/contact)
