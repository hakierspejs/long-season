CREATE TABLE users (
    userID TEXT NOT NULL,
    userNickname TEXT UNIQUE NOT NULL,
    userPassword BLOB NOT NULL,
    userPrivate int,
    PRIMARY KEY(userID)
);

CREATE TABLE devices (
    deviceID TEXT PRIMARY KEY,
    deviceOwnerID TEXT NOT NULL,
    deviceTag TEXT,
    deviceMAC BLOB NOT NULL,
    FOREIGN KEY(deviceOwnerID) REFERENCES users(userID)
);

CREATE TABLE otp (
    otpID TEXT PRIMARY KEY,
    otpName TEXT NOT NULL,
    otpSecret BLOB NOT NULL,
    otpOwnerID TEXT NOT NULL,
    FOREIGN KEY(otpOwnerID) REFERENCES users(userID)
);

CREATE TABLE recovery (
    recoveryID TEXT PRIMARY KEY,
    recoveryName TEXT NOT NULL,
    recoveryOwnerID TEXT NOT NULL,
    FOREIGN KEY(recoveryOwnerID) REFERENCES users(userID)
);

CREATE TABLE recoveryCodes (
    recoveryCodesCode TEXT NOT NULL,
    recoveryCodesID TEXT NOT NULL,
    FOREIGN KEY(recoveryCodesID) REFERENCES recovery(recoveryID)
);

-- Remove all recovery codes when user attempt to
-- remove recovery two factor method.
CREATE TRIGGER cleanRecoveryCodes
    AFTER DELETE ON recovery
BEGIN
    DELETE FROM
        recoveryCodes
    WHERE
        recoveryCodes.recoveryCodesID = old.recoveryID;
END;

-- Remove recovery codes metadata after usage of last recovery
-- code.
CREATE TRIGGER cleanRecovery
    AFTER DELETE ON recoveryCodes
WHEN
    (SELECT
        count(*)
    FROM
        recoveryCodes
    WHERE
        recoveryCodes.recoveryCodesID = old.recoveryCodesID) = 0
BEGIN
    DELETE FROM
        recovery
    WHERE
        recovery.recoveryID = old.recoveryCodesID;
END;
