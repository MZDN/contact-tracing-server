# Contact Tracing Cloud API

Status: **Implementation In Progress**

This Contact Tracing API is used for _Privacy-Preserving Distributed Contact Tracing_, as used in Wolk's **Experimental** Contact Tracing Implementations:
* [contact-tracing-ios](https://github.com/wolkdb/contact-tracing-ios) - iOS Client (pure native Swift)
* [contact-tracing-android](https://github.com/wolkdb/contact-tracing-android) - Android Client (pure native Kotlin)
Wolk is making this server code complete to support application developers (commercial or non-commercial), governments and health organizations.  

If you are interested in getting an API key, email sourabh@wolk.com

If you are interested in contributing code to any contact-tracing, fork the repo and submit a pull request


## Active Endpoint
* API Endpoint: (active) https://api.wolk.com
* API Documentation in Postman: https://documenter.getpostman.com/view/10811660/SzYeww4L

## Team Leads
* Sourabh Niyogi (contact-tracing-android, contact-tracing-server), sourabh@wolk.com
* Mayumi Matsumoto (contact-tracing-server)
* Michael Chung (contact-tracing-server)
* Rodney Witcher (contact-tracing-ios)

## Background

On April 10, 2020, Apple and Google announced detailed protocol specifications for Privacy-Preserving Contact Tracing, making all earlier contact tracing protocols immaterial and enabling constructive activity on life-saving application development:

[Apple 4/10/20 Announcement](https://www.apple.com/newsroom/2020/04/apple-and-google-partner-on-covid-19-contact-tracing-technology/)
* [Contact Tracing - Bluetooth Specification](https://covid19-static.cdn-apple.com/applications/covid19/current/static/contact-tracing/pdf/ContactTracing-BluetoothSpecificationv1.1.pdf)
* [Contact Tracing - Cryptography Specification](https://covid19-static.cdn-apple.com/applications/covid19/current/static/contact-tracing/pdf/ContactTracing-CryptographySpecification.pdf)
* [Contact Tracing - Framework API](https://covid19-static.cdn-apple.com/applications/covid19/current/static/contact-tracing/pdf/ContactTracing-FrameworkDocumentation.pdf)

[Google 4/10/20 Announcement](https://blog.google/inside-google/company-announcements/apple-and-google-partner-covid-19-contact-tracing-technology):
* [Privacy-safe contact tracing using Bluetooth Low Energy](https://blog.google/documents/57/Overview_of_COVID-19_Contact_Tracing_Using_BLE.pdf)
* [Contact Tracing Bluetooth Specification](https://blog.google/documents/58/Contact_Tracing_-_Bluetooth_Specification_v1.1_RYGZbKW.pdf)
* [Contact Tracing Cryptography Specification](https://blog.google/documents/56/Contact_Tracing_-_Cryptography_Specification.pdf)
* [Android Contact Tracing API](https://blog.google/documents/55/Android_Contact_Tracing_API.pdf)

### BigTable Setup

You can set up your API endpoint in a Google Cloud project of your own.

1. Set up your BigTable instance in a Google Cloud project such as
```
project = yourGCProject
instance = yourBTInstance
```
and use `cbt` (see [Quickstart](https://cloud.google.com/bigtable/docs/quickstart-cbt) to create a BigTable `report` with a column family `report`:
```
cbt createtable report
cbt createfamily report report
cbt ls
cbt ls report
cat ~/.cbtrc
```
add `project` and `instance` to `conf/fmpk.conf`
```
        "bigtableProject": "yourGCProject",
        "bigtableInstance": "yourBTInstance"
```
2. Getting your SSL Certs (for `example.com`) into `backend` package
3. Set up a DNS entry (`contact-tracing.example.com`) that matches and running `bin/contact-tracing`
4. Build the `findmypk` server and run it!
```
$ make contact-tracing
go build -o bin/contact-tracing
Done building contact-tracing!  Run "bin/contact-tracing" to launch diagnosis Server.
$ bin/contact-tracing
Contact Tracing Diagnosis Server Listening on port 443...
```

## Test
Run a test with: (under construction)
```
# go test -run TestContactTracingSimple
...
PASS
```
