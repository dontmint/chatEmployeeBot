CREATE TABLE employee (
    employeeId SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    storeId INTEGER NOT NULL,
    salaryRate INTEGER NOT NULL,
    active INTEGER NOT NULL,
    FOREIGN KEY(storeId) REFERENCES store(storeId)
);

CREATE TABLE store (
    storeId SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    active INTEGER NOT NULL
);

CREATE TABLE workRecord (
    employeeId INTEGER NOT NULL,
    time TEXT NOT NULL,
    year TEXT NOT NULL,
    month TEXT NOT NULL,
    day TEXT NOT NULL,
    checkin INTEGER NOT NULL,
    checkout INTEGER NOT NULL
);

CREATE TABLE storeManager (
    managerID SERIAL PRIMARY KEY,
    storeID INTEGER NOT NULL,
    active INTEGER NOT NULL,
    FOREIGN KEY(storeID) REFERENCES store(storeId)
);

CREATE TABLE workTracking (
    employeeId INTEGER NOT NULL,
    storeId INTEGER NOT NULL,
    working INTEGER NOT NULL,
    PRIMARY KEY (employeeId, storeId)
);

CREATE INDEX employee_storeid_index ON employee (storeId);
CREATE INDEX storemanager_storeid_index ON storeManager (storeID);
