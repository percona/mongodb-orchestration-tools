# mongodb_tools

Go-based tools for the DCOS 'percona-mongo' service

- **mongodb-executor**: wrapper tool for 'mongod' and executing tasks on the local mongod/mongos container
- **mongodb-healthcheck**: tool for running DCOS health and readiness checks on a MongoDB task
- **mongodb-initiator**: tool for initiating the replica set and adding system users
- **mongodb-watchdog**: daemon to monitor dcos pod status and manage mongodb replica set membership

## Build

1. Install go1.8+ and 'make'
2. Run 'make' in git directory
3. Find binaries in 'bin' directory

## Contact
- Tim Vaillancourt - [Github](https://github.com/timvaillancourt) [Email](mailto:tim.vaillancourt@percona.com)
- Percona - [Twitter](https://twitter.com/Percona) [Contact Page](https://www.percona.com/about-percona/contact)
