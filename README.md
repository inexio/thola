# Thola

<img align="right" src="https://raw.githubusercontent.com/inexio/thola/main/doc/logo.png" width="256" alt="Thola">

[![Go Report Card](https://goreportcard.com/badge/github.com/inexio/thola)](https://goreportcard.com/report/github.com/inexio/thola)
[![GitHub code style](https://img.shields.io/badge/code%20style-uber--go-brightgreen)](https://github.com/uber-go/guide/blob/master/style.md)
[![GitHub license](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/inexio/thola/blob/main/LICENSE)
[![GitHub branch checks state](https://img.shields.io/github/checks-status/inexio/thola/main)](https://github.com/inexio/thola/actions)
[![GoDoc doc](https://img.shields.io/badge/docs-online-brightgreen)](https://docs.thola.io)

## Description

A tool for monitoring network devices written in Go.
It features a check mode which complies with the [monitoring plugins development guidelines](https://www.monitoring-plugins.org/doc/guidelines.html) and is therefore compatible with Nagios, Icinga, Zabbix, Checkmk, etc.

## Installation

You can download the latest compiled version for your platform under the "Releases" tab or build it yourself:

    git clone https://github.com/inexio/thola.git
    cd thola
    go build
    
**Note: This requires Go 1.16 or newer**

If you also want to build the client binary, which can be used for sending requests to a running Thola API, use the following build command:
   
    go build --tags client -o thola-client

## Features

Thola currently has three main modes of operation with various subcommands:

- `identify` automatically identifies the device and outputs its vendor, model and other properties.
- `read` reads out values and statistics of the device.
    - `read available-components` returns the available components for the device.
    - `read count-interfaces` counts the interfaces.
    - `read cpu-load` returns the current cpu load of all CPUs.
    - `read disk` reads storage utilization.
    - `read hardware-health` reads hardware health information like temperatures and fans.
    - `read high-availability` reads out the high availability status of a device.
    - `read interfaces` outputs the interfaces with several values like error counters and statistics.
    - `read sbc` reads out SBC specific information.
    - `read memory-usage` reads out the current memory usage.
    - `read server` outputs server specific information like users and process count.
    - `read ups` outputs the special values of a UPS device.
- `check` performs checks that can be used in monitoring systems. Output is by default in check plugin format.
    - `check cpu-load` checks the average CPU load of all CPUs against given thresholds and outputs the current load of all CPUs as performance data.
    - `check disk` checks the free space of storages.
    - `check hardware-health` checks the hardware-health of a device.
    - `check high-availability` checks the high availability status of a device.
    - `check identify` compares the device properties with given expectations.
    - `check interface-metrics` outputs performance data for the interfaces, including special values based on the interface type (e.g. Radio Interface).
    - `check memory-usage` checks the current memory usage against given thresholds.
    - `check sbc` checks an SBC device and outputs metrics for each realm and agent as performance data.
    - `check server` checks server specific information.
    - `check snmp` checks SNMP reachability.
    - `check ups` checks if a UPS device has its main voltage applied and outputs additional performance data like battery capacity or current load, and compares them to optionally given thresholds.
    - `check thola-server` checks reachability of a Thola API.

## Quick Start

Use the `identify` mode to automatically discover some properties of a network device.
    
    $ thola identify
    
    Usage:
      thola identify [host] [flags]
Specify the address of the network device in the `[host]` argument.
The `--format` flag modifies the format of the output. `--format pretty` is set by default and is useful when reading the output manually. Other options are `json` and `xml`.

    $ thola identify 10.204.2.90
    
    Device: 
      Class: ceraos/ip10
      Properties: 
        Vendor: Ceragon
        Model: IP-10
        SerialNumber: 00:0A:25:25:77:67
        OSVersion: 2.9.25-1
Next we want to print the interfaces of the network device and their relevant data. We use the `read interfaces` command for this.

    $ thola read interfaces 10.204.2.90
    
    Interfaces: [8] 
      IfIndex: 1
      IfDescr: Radio Interface #0
      IfType: sonet
      IfMtu: 2430
      IfSpeed: 367000
      ...
      
      IfIndex: 5001
      IfDescr: Ethernet #7
      IfType: ethernetCsmacd
      IfMtu: 1548
      IfSpeed: 10000000
      IfPhysAddress: 00:0A:25:27:57:1E
      IfAdminStatus: up
      IfOperStatus: down
      ...

## API Mode

Thola can be executed as a REST API. You can start the API using the `api` command:

    $ thola api
     ______   __  __     ______     __         ______   
    /\__  _\ /\ \_\ \   /\  __ \   /\ \       /\  __ \  
    \/_/\ \/ \ \  __ \  \ \ \/\ \  \ \ \____  \ \  __ \ 
       \ \_\  \ \_\ \_\  \ \_____\  \ \_____\  \ \_\ \_\
        \/_/   \/_/\/_/   \/_____/   \/_____/   \/_/\/_/
    
    ⇨ http server started on [::]:8237
    
For sending requests to the Thola API you can use the Thola client. When executing the Thola client you can specify the address of the API with the `--target-api` flag.

    $ thola-client identify 10.204.2.90 --target-api http://192.168.10.20:8237 
    
    Device: 
      Class: ceraos/ip10
        Properties: 
          Vendor: Ceragon
          Model: IP-10
          SerialNumber: 00:0A:25:25:77:67
          OSVersion: 2.9.25-1
        
You can find the full API documentation on our [SwaggerHub](https://app.swaggerhub.com/apis-docs/thola/thola/1.0.0).

## Supported Devices

We support a lot of different devices and hope for your contributions to grow our device collection. Some examples are:

- Cisco
- Juniper
- Huawei
- Nokia/ISAM
- Ceragon
- Brocade
- Edgecore
- ...

Basic interface readout is supported for every device.

## Supported Protocols

Currently we mostly work with SNMP, but already provide basic features for HTTP(S).
We plan to support more protocols like telnet, SSH and more.

## Tests

You can run our test located in the `test` directory with the `go test` command if you have Docker and Docker Compose installed. 

If you want to add your own devices  to the tests you can put your SNMP recordings in the `testdata/devices` folder.
After that you just need to run the script located in `create_testdata` to create the expectation files and your devices are included in the testsuite!

## Contribution

We are always looking forward to your ideas and suggestions.

If you want to help us please make sure that your code is conform to our coding style.

Happy coding!
