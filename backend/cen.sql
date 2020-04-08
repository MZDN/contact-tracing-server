DROP TABLE IF EXISTS CENReport;
DROP TABLE IF EXISTS CENSymptom;

CREATE TABLE `CENSymptom` (
	   `reportID` varchar(64) DEFAULT "",
	   `symptomID` int,
	   `reportMimeType` varchar(64) DEFAULT "",
	   `reportTS` int,
	   `storeTS` int,
	   PRIMARY KEY (`reportID`, `symptomID`),
	   KEY (`reportTS`)
);

CREATE TABLE `CENReport` (
   `hashedPK`  varchar(64),
   `encodedMsg` varchar(512),
   `reportTS` int,
   `prefixHashedPK` varchar(6),
   PRIMARY KEY(`hashedPK`),
   KEY(`prefixHashedPK`)
);




