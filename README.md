# OpenChoreo Sample Workloads

Welcome to the official repository of sample workloads for **[OpenChoreo](https://github.com/openchoreo/openchoreo)**! 

This repository provides a curated collection of examples to help you get started with OpenChoreo, the open-source internal developer platform. These samples demonstrate various deployment patterns and best practices for building, deploying, and managing your applications on the platform.

## Table of Contents

- [Overview](#-overview)
- [Directory Structure](#-directory-structure)
- [Prerequisites](#-prerequisites)
- [How to Use](#-how-to-use)
- [Contributing](#-contributing)
- [License](#-license)


## Overview

These workloads demonstrate the key capabilities of OpenChoreo’s built-in CI system, including:

- **Building from source** using Cloud Native Buildpacks or Dockerfiles.
- **Custom configuration** for environments, pipelines, and secure services.
- **Real-world use cases** like scheduled tasks, services, and web apps.

Each sample includes:
- Source code (Go, Python, React, etc.)
- A `workload.yaml` file describing endpoints and connections


## Directory Structure


Each folder follows the convention `<component-type>-<language>-<name>`, with examples such as:

```bash
├── service-go-reading-list/  
├── service-go-greeter/              
├── webapp-react-nginx/         
└── <component-type>-<language>-<name>/ 
```


## How to Use

Each sample includes a `workload.yaml` file, which is the declarative manifest for deploying applications on OpenChoreo.

Refer to the [Samples](https://github.com/openchoreo/openchoreo) in OpenChoreo repository to see how these workload examples are used in real CI workflows.


## Contributing

We welcome contributions! If you have a sample workload that demonstrates a new pattern or use case, please feel free to open a pull request.



## License

This project is licensed under the Apache 2.0 License - see the **[LICENSE](LICENSE)** file for details.
