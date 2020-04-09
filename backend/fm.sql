DROP TABLE IF EXISTS FMReport;
DROP TABLE IF EXISTS FMSymptom;

CREATE TABLE `FMSymptom` (
	   `reportID` varchar(64) DEFAULT "",
	   `symptomID` int,
	   `reportMimeType` varchar(64) DEFAULT "",
	   `reportTS` int,
	   `storeTS` int,
	   PRIMARY KEY (`reportID`, `symptomID`),
	   KEY (`reportTS`)
);

CREATE TABLE `FMReport` (
   `hashedPK`  varchar(64),
   `encodedMsg` varchar(512),
   `reportTS` int,
   `prefixHashedPK` varchar(6),
   PRIMARY KEY(`hashedPK`),
   KEY(`prefixHashedPK`)
);
