DROP TABLE IF EXISTS "reg_provinces";
CREATE TABLE "reg_provinces" (
  id CHAR(2) PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);


DROP TABLE IF EXISTS "reg_regencies";
CREATE TABLE "reg_regencies" (
  id CHAR(4) PRIMARY KEY,
  province_id CHAR(2) NOT NULL,
  name VARCHAR(255) NOT NULL,
  CONSTRAINT fk_province FOREIGN KEY (province_id) REFERENCES reg_provinces(id)
);


DROP TABLE IF EXISTS "reg_districts";
CREATE TABLE "reg_districts" (
  id CHAR(6) PRIMARY KEY,
  regency_id CHAR(4) NOT NULL,
  name VARCHAR(255) NOT NULL,
  CONSTRAINT fk_regency FOREIGN KEY (regency_id) REFERENCES reg_regencies(id)
);


DROP TABLE IF EXISTS "reg_villages";
CREATE TABLE "reg_villages" (
  id CHAR(10) PRIMARY KEY,
  district_id CHAR(6) NOT NULL,
  name VARCHAR(255) NOT NULL,
  CONSTRAINT fk_district FOREIGN KEY (district_id) REFERENCES reg_districts(id)
);
