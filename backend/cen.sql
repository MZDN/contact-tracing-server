DROP TABLE IF EXISTS CENKeys;
DROP TABLE IF EXISTS CENReport;
DROP TABLE IF EXISTS CENSymptom;
DROP TABLE IF EXISTS CENSymptomType;
DROP TABLE IF EXISTS CENStatus;
DROP TABLE IF EXISTS CENStatusType;
DROP TABLE IF EXISTS CENSymptomDictionary;

-- reportID is uniq per entire record or per user?
CREATE TABLE `CENKeys` (
   `cenKey`   varchar(32) DEFAULT "", 
   `reportID` varchar(64) DEFAULT "",
   `reportTS` int,
   PRIMARY KEY (`cenKey`, `reportID`),
   KEY (`reportID`),
   KEY (`reportTS`),
   KEY (`cenKey`)
);

CREATE TABLE `CENReport` (
   `reportID` varchar(64) DEFAULT "",
   `report`     varchar(4000) DEFAULT "",
   `reportMimeType` varchar(64) DEFAULT "",
   `reportTS` int,
   `storeTS` int,
   PRIMARY KEY (`reportID`),
   KEY (`reportTS`)
);

CREATE TABLE `CENSymptom` (
   `reportID` varchar(64) DEFAULT "",
   `symptomID` int,
   `reportMimeType` varchar(64) DEFAULT "",
   `reportTS` int,
   `storeTS` int,
   PRIMARY KEY (`reportID`, `symptomID`),
   KEY (`reportTS`)
);

/*
symptomID 	symptom
1		Tiredness
2		Dry cough
3		Muscle aches
....
*/

CREATE TABLE `CENSymptomType` (
   `symptomID` int,
   `symptom` varchar(32) DEFAULT "",
   PRIMARY KEY (`symptomID`)
);

/* not overwrite status. you can trace the status by CENKeys' reportTS */
CREATE TABLE `CENStatus` (
   `reportID` varchar(64) DEFAULT "",
   `statusID` int,
   `reportTS` int,
   `storeTS` int,
   PRIMARY KEY(`statusID`)
);

/* 
statusID	status
0		noreport
1		positive
2		negative
3		recovered
*/
CREATE TABLE `CENStatusType` (
   `statusID`  int,
   `status` varchar(32),
   PRIMARY KEY(`statusID`)
);

CREATE TABLE `CENSymptomDictionary` (
   `symptomID` int,
   `symptomWord` varchar(32)
);

CREATE TABLE `CENReport` {
   `hashedPK`  varchar(32),
   `encodedMsg` varchar(256),
   `reportTS` int,
   `prefixHashedPK` varchar(24),
   PRIMARY KEY(`hashedPK`),
   KEY(`prefixHashedPK`)
};




