DROP TABLE IF EXISTS FMReport;

CREATE TABLE `FMReport` (
   `hashedPK`  varchar(64),
   `encodedMsg` varchar(512),
   `reportTS` int,
   `prefixHashedPK` varchar(6),
   PRIMARY KEY(`hashedPK`),
   KEY(`prefixHashedPK`)
);
