# Find My PK (FMPK) Protocol v0.1
Inspired by Apple’s _Find My_ Protocol

Gary Belvin, gdbelvin@gmail.com
README based on [2020 April 5](https://docs.google.com/document/d/1jFOic0--h1Le5x44iwj3jY0k23nUuEL9gGgPSE0H-hs/edit#)


Status: Implementation In Progress

Implementation by [Wolk](https://wolk.com) under COVID License:
* `wolkdb/findmypk-server` - Go Server using Google BigTable
* `wolkdb/findmypk-ios` - iOS Client (pure native Swift)
* `wolkdb/findmypk-android` - Android Client (pure native Kotlin)

Contributors:
* Sourabh Niyogi (findmypk-android, findmypk-server), sourabh@wolk.com
* Mayumi Matsumoto (findmypk-server)
* Michael Chung (findmypk-server)
* Rodney Witcher (findmypk-ios)

Contributions sought on:
* iOS-iOS background BLE Scan (submit PR)
* Security analyses (submit PR)

## Objectives
* Enable users to notify the people they were physically proximate to in the last two weeks if they develop symptoms for CoVID19.
* Do so in a privacy preserving way that does not leak their contacts, or health status to anyone other than the intended recipients (the people they were physically proximate to)
* Compatibility with existing protocols such as Apple’s Find My protocol
* Interoperability with the TCN protocol

## Security Goals
* Server Privacy: An honest-but-curious server should not learn information about any user's location or contacts.
* Source Integrity: Users cannot send reports to users they did not come in contact with or on behalf of other users.
* Broadcast Integrity: Users cannot broadcast TCNs they did not generate.
* No Passive Tracking: A passive adversary monitoring Bluetooth connections should not be able to learn any information about the location of users who do not send reports.
* Receiver Privacy: Users who receive reports do not reveal information to anyone.
* Reporter Privacy: Users who send reports do not reveal information to users they did not come in contact with, and reveal only the time of contact to users they did come in contact with. Note that in practice, the timing alone may still be sufficient for their contact to learn their identity (e.g., if their contact was only around one other person at the time).

## Protocol

### Receivers

1. Every 15 minutes, or as often as a phone’s MAC address changes
   * Generate an ephemeral public, private key pair `(pk, sk)`
On iOS it is not possible to tell how often the MAC address changes, so the best we can do is rotate this key faster than the MAC rotation. [Reference](https://petsymposium.org/2019/files/papers/issue3/popets-2019-0036.pdf)

2. Broadcast `pk, sig, m` over Low Energy Bluetooth

    * The public key `pk` acts like a random, untraceable, contact event number.
    * The sig is over `m` which may or may not reveal symptoms, diseases, and health authority certification.

    *Broadcast method*:
    * Use Service UUID (but no Service Data)
    * Use Characteristic Data field (max 512 bytes) to broadcast:
`u8 | PublicKey | u8 | Signature | u8 | memo`
where each u8 is the number of bytes for the attribute immediately following it.  The broadcaster keep the private keys around in local storage for 2 weeks to decrypt messages retrieved later.

## Reporters
1. Record all observed public keys `(pk,sig,memo)` (after Signature verification) as they are observed and store them for at least two weeks  

2. In the event that an individual develop symptoms for CoVID 19 or has a positive test
a. Construct a message “m” using this protobuf:
```
message FindMyPKMemo {
  enum ReportType {
   SELF_REPORTED       = 0;
   CERTIFIED_INFECTION = 1;
  }
  ReportType     reportType = 1;
  int32 diseaseID = 2; // 0=Healthy, 1=Unhealthy/Unknown, 2= COVID-19
  repeated int32 symptomID = 3;  
  // Could be used with no key rotation while someone is quarantined
  bytes publicHealthAuthorityPublicKey = 4;
  bytes signature = 5;  
}
```
b. Encrypt the message `m`:

* Generate an ephemeral public key `pk_b`
* Combine `pk_b` with `pk_a` using Diffie-Hellman to produce a session secret (`ss`)
* Encrypt the message using AES_GCM
Ciphertext = IV, EncMsg, MAC = AES_GCM(ss, m)

Upload a vector of tuples `[]((H(pk_a), pk_b, AES_GCM(DH(pk_a, pk_b), m))`
By not uploading or sharing `pk`, we preserve the notion that knowledge of `pk` is evidence of having observed `pk`.

## Servers

1. Store tuples in buckets indexed by a fixed length prefix `H(pk)`,
This enables users to query buckets without revealing the full `H(pk)` they own or have observed, a form of private set intersection.

## Inter-Server Sync
Because these tuples do not contain private information and are self verifying, the server’s dataset can be made public and be synced with other servers without permissioning.

## Query
Receivers periodically query the server for messages that may have been sent to the public keys they broadcast.

1. The receiver sends a vector of prefixes of H(pk) to the server.
* 1 prefix (the first n  bits of `H(pk)`) per 15 minute window the user would like to query for contact overlap.
* This prevents the server from knowing the exact H(pk) being queried, and provides k anonymity / plausible deniability about whether there is a match.
2. The server responds with all `(H(pk), Enc(pk, m))` tuples in the requested buckets.
3. The receiver discards messages that are not for H(pk)
4. The receiver validates messages.
5. At the end of the query, the receiver will know:
a. The number of times the receiver was exposed in the past 2 weeks.
b. The times, locations, and other metadata the receiver may have recorded along with the pks at the time they were broadcast.


## Performance Costs

### Assumptions

| # | Stat | Notes/Assumptions|
|----------|---------|------------------------------|
| 1.00E+06 | Reports | https://ourworldindata.org/grapher/total-cases-covid-19 |
| 1.00E+09 | Receivers | https://datareportal.com/global-digital-overview
| 1344 | FMPK Per Report | Assumes a person saw a unique FMPK ever 15 min for 2 weeks
| 1344 | Prefix Matches Per Query | Assumes the receiver was near an infected person every 15 min for 2 weeks
| 1 |  Days between queries |
| 32/65 |  Public Key Bytes  | ECDSA-256 |
| 32 | H(pk) | SHA-256 |
| 29 | Enc(pk, msg) | Len(AES_GCM(DH(pk_a, pk_b), msg))|
| 93 | Bytes Per FMPK report | H(pk) + pk_b + Enc(pk, msg) |


## Bandwidth Costs

| Prefix Size | K Anonymity | Bucket Size (Kb) |  Query (Kb) | Response (Mb) | Egress (Mbs) |
|-------------|-------------|------------------|-------------|------------------|--------------|
| 18 | 5,127 | 465.6 | 24 |  611.1 | 117,890 |
| 20 | 1,282 | 116.4 | 26 | 152.8 | 29,472 |
| 24 |  80 |  7.3 | 32 | 9.5 |  1,842 |
| 30 |  1  |  0.1 | 39 | 0.1 | 29 |

K Anonymity in this table is "per FMPK". This does not provide K anonymity for the user because each user will have a unique pattern of queries across multiple FMPK buckets.

## Storage Costs

|#      | Storage          |
|-------|----------------- |
| 106   | Kb Per Report    |
| 101   | Gb Total Storage |

## Compatibility

### With Apple’s Find My Protocol

Apple is already broadcasting ephemeral public keys over low energy bluetooth.

By observing, storing and encrypting messages with these public keys, a service could collaborate with Apple to do the lookup protocol.  Apple’s cooperation here would be needed since they control the app that has access to the private keys that this protocol would be encrypting messages to.

In Apple’s Find My protocol, broadcast is “opt in”, listening is “opt out”

Integration with Apple’s protocol may not be be as desirable a design goal as I initially thought, both because of the opt-in issue, but also because of the unlikely possibility of gaining access to the private keys associated with the Find My Protocol.

### With TCN protocol

The public key can be treated like the temporary contact number (TCN) because it is ephemeral and changing every 15 minutes.

The present TCN protocol, however, reveals TCNKeys directly to the server during the reporting phase. and the query phase, breaking the ability to reason accurately about whether TCNs were observed or queried.

It could be remedied by modifying the TCN protocol to upload (H(TCN), HMAC(TCN, 1)) during the reporting phase, shielding TCN from the server.  This is not possible because TCNv3 downloads batches of TCNs via TCNKey’s to make the protocol tractable while retaining full privacy.

## Security and Privacy Comparison to TCN

### Server Privacy: An honest-but-curious server should not learn information about any user's location or contacts.

TCN or H(TCN) is a random ID that changes frequently enough that it should not be linkable across times and locations.  

Malicious parties could go around geotagging observed TCNs, effectively adding a map to a report that they were physically present to observe.


### Source Integrity: Users cannot send reports to users they did not come in contact with or on behalf of other users.

By uploading a proof of knowledge of TCN in the form of a) Enc(TCN,1) or b HMAC(TCN, 1), we prevent others from being able to generate reports without directly observing TCN for themselves.


### Broadcast Integrity: Users cannot broadcast TCNs they did not generate.

In the present form, malicious parties could rebroadcast TCNs.

This could be mitigated by encrypting the time of the observed TCN in the report. This would give the receiver additional opportunities to validate the TCN, and prevent rebroadcasting at different times. If preventing rebroadcast at the same time, but in different locations is a goal, the location could be included in the message as well.

###  No Passive Tracking: A passive adversary monitoring Bluetooth connections should not be able to learn any information about the location of users who do not send reports.

Users who do not send reports do not upload information other than:
* The prefix of `H(pk)`
* New unlinkable TCNs over bluetooth

Physically co-present passive trackers on bluetooth will be able to associate times and locations with the TCNs they observe.

### Receiver Privacy: Users who receive reports do not reveal information to anyone.

By uploading many H(pk) prefixes, users uniquely identify themselves.
If this is a significant problem, fancier cryptographic private set intersection protocols can be used, but these become complex and computationally expensive.

Users can break linkability between requests, however, by making requests in non-overlapping batches. Or by making use of a mix net or a proxy to make the requests.

*Update:* This is not robust. Because users can be individually identified through their queries, and, to a rough extent (as obfuscated by PIR), can discover what a user’s contacts are -- a significant privacy issue

### Reporter Privacy: Users who send reports do not reveal information to users they did not come in contact with, and reveal only the time of contact to users they did come in contact with. Note that in practice, the timing alone may still be sufficient for their contact to learn their identity (e.g., if their contact was only around one other person at the time).

Reporters will only send information to TCNs they observed. If TCN rebroadcast is prevented, this is a strong indicator of co-presence. Receivers can associate particular TCNs to the time and location they were at when they broadcast that TCN.

## Abuse

### Warning fatigue

is real and related to both warning frequency and user perceptions of accuracy.

Abusive users can negatively affect warning accuracy by generating reports to every TCN they observe. Abuse need not be malicious, this could occur at-scale if a non trivial percentage of the user population reports out of worry or a desire to be overly helpful “just in case”.

Warning fatigue could also set in if the product is wildly successful, and a significant percentage of the population is indeed sick or at-risk.

### Denial of Service
Abusive reporters can overwhelm the service by generating large numbers of artificial TCNs and uploading them. Fortunately, the service will expire them in 2 weeks, but we may want to cap, both the number of TCNs in an individual report, and the number of times a particular IP can generate reports.

Because this protocol preserves the privacy of the reporter, it is also possible to encrypt a large number of messages to the same TCN.  We may want to limit the number of messages encrypted to a particular TCN at the service layer to a small number between 1 and 10.

# Find My Public Key (FMPK) API
This Find My Public Key API is used for _Privacy-Preserving Distributed Contact Tracing_, as used in FMPK Apps:
* [findmypk-ios](https://github.com/wolkdb/findmypk-ios)
* [findmypk-android](https://github.com/wolkdb/findmypk-android)
* other applications following [FMPK Protocols](https://github.com/wolkdb/findmypk-server)
TCN Coalition (https://tcn-coalition.org).
The TCN Coalition mission is to reduce transmission of disease, by developing applications and protocols that support contact tracing without loss of privacy (minimal identifiable information).  The Find My Public Key (FMPK) Protocol achieves this goal by combining Bluetooth Low Energy with servers holding no PII data and clients doing all matching.
The flow is as follows:
1. iOS/Android Apps broadcast `(pk, sig, memo)` using Bluetooth Low Energy (BLE) in specific Service / Characteristic ID
2. FMPK App records neighboring `pk`  in close physical proximity to the user.
3. Users submit symptom / infection reports in their application to a FMPK API endpoint, resulting in a POST to `/report` endpoint with `Report`  
4. Apps poll periodically (hourly / daily) to `/query` for recent `Report` and matches them to `pk`s the user has observed, matching them locally.
## Active Endpoint
* API Endpoint: (active) https://api.wolk.com
* API Documentation in Postman: TBD
### BigTable Setup
1. Set up your BigTable instance in a Google Cloud project such as
```
project = findmypk-us-west1
instance = findmypk
```
and use `cbt` (see [Quickstart](https://cloud.google.com/bigtable/docs/quickstart-cbt) to create a BigTable `report` with a column family `report`:
```
cbt createtable report
cbt createfamily report report
cbt ls
cbt ls report
cat ~/.cbtrc
```
2. Getting your SSL Certs (for `example.com`) into `backend` package
3. Set up a DNS entry (`findmypk.example.com`) that matches and running `bin/findmypk`
4. Build the `findmypk` server and run it!
```
$ make findmypk
go build -o bin/findmypk
Done building FindMyPk!  Run "bin/findmypk" to launch findmypk Server.
$ bin/findmypk
FindMyPk Server Listening on port 443...
```
## Test
Run a test with: (under construction)
```
# go test -run TestFMPKSimple
...
PASS
```
