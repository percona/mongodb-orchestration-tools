# dcos-mongo-tools

Go-based tools for the [DC/OS 'percona-mongo' service](https://docs.mesosphere.com/services/percona-mongo/)

- **mongodb-executor**: tool for executing tasks on the local mongod/mongos container
- **mongodb-controller**: tool for controlling the replica set initiation and adding system MongoDB users
- **mongodb-healthcheck**: tool for running DC/OS health and readiness checks on a MongoDB task
- **mongodb-watchdog**: daemon to monitor dcos pod status and manage mongodb replica set membership

## Build

1. Install go1.8+ and 'make'
2. Run 'make' in git directory
3. Find binaries in 'bin' directory

## Contact
- Tim Vaillancourt - [Github](https://github.com/timvaillancourt) [Email](mailto:tim.vaillancourt@percona.com)
- Percona - [Twitter](https://twitter.com/Percona) [Contact Page](https://www.percona.com/about-percona/contact)
