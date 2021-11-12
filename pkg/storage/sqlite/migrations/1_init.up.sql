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
    CONSTRAINT fkDevices
        FOREIGN KEY(deviceOwnerID)
        REFERENCES users(userID)
        ON DELETE CASCADE
);

CREATE TABLE otp (
    otpID TEXT PRIMARY KEY,
    otpName TEXT NOT NULL,
    otpSecret BLOB NOT NULL,
    otpOwnerID TEXT NOT NULL,
    CONSTRAINT fkOtp
        FOREIGN KEY(otpOwnerID)
        REFERENCES users(userID)
        ON DELETE CASCADE
);

CREATE TABLE recovery (
    recoveryID TEXT PRIMARY KEY,
    recoveryName TEXT NOT NULL,
    recoveryOwnerID TEXT NOT NULL,
    CONSTRAINT fkRecovery
        FOREIGN KEY(recoveryOwnerID)
        REFERENCES users(userID)
        ON DELETE CASCADE
);

CREATE TABLE recoveryCodes (
    recoveryCodesCode TEXT NOT NULL,
    recoveryCodesID TEXT NOT NULL,
    CONSTRAINT fkRecoveryCodes
        FOREIGN KEY(recoveryCodesID)
        REFERENCES recovery(recoveryID)
        ON DELETE CASCADE
);

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
