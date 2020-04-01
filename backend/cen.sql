DROP TABLE IF EXISTS CENKeys;
DROP TABLE IF EXISTS CENReport;

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

/*
reportID is not uniq. 
*/
CREATE TABLE `CENReport_M` (
   `reportID` varchar(64) DEFAULT "",
   `symptomID` int,
   `reportMimeType` varchar(64) DEFAULT "",
   `reportTS` int,
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

CREATE TABLE `CENSymptom` (
   `symptomID` int,
   `symptom` varchar(32) DEFAULT "",
   `reportMimeType` varchar(64) DEFAULT "",
   `reportTS` int, 
   PRIMARY KEY (`symptomID`),
   KEY (`reportTS`)
);

/* not overwrite status. you can trace the status by CENKeys' reportTS */
CREATE TABLE `CENStatus` (
   `reportID` varchar(64) DEFAULT "",
   `statusID` int,
   PRIMARY KEY(`statusID`)
);

/* 
statusID	status
0		noreport
1		positive
2		negative
3		recovered
*/
CREATE TABLE `CENStatyeType` (
   `statusID`  int,
   `status` varchar(32),
   PRIMARY KEY(`statusID`)
);



