import { withErr } from "/static/js/utils.js";

class AuthorizationRequiredError extends Error {
  constructor(message) {
    super(message);
    this.name = "AuthorizationRequiredError";
  }
}

class HTTPError extends Error {
  constructor(message) {
    super(message);
    this.name = "HTTPError";
  }
}

async function who() {
  let [res, err] = await withErr(fetch("/who", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  }));
  if (err) {
    return [null, err];
  }
  if (!res.ok) {
    let authErr = new AuthorizationRequiredError("User is not authorized.");
    return [null, authErr];
  }
  let [parsed, jsonErr] = await withErr(res.json());
  if (jsonErr) {
    return [null, err];
  }
  return [parsed, null];
}

async function updatePassword(userID, { oldPass, newPass }) {
  let [res, err] = await withErr(fetch(`/api/v1/users/${userID}/password`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    redirect: "follow",
    body: JSON.stringify({
      "old": oldPass,
      "new": newPass,
    }),
  }));
  if (err) {
    return [null, err];
  }
  return [res, null];
}

async function optionsOTP() {
  let [res, errOptions] = await withErr(fetch("/api/v1/twofactor/otp/options", {
    method: "GET",
    redirect: "follow",
    credentials: "include",
  }));
  if (errOptions) {
    return [null, err];
  }
  let [jsonRes, errJson] = await withErr(res.json());
  if (errJson) {
    return [null, errJson];
  }
  return [jsonRes, null];
}

async function newOTP(body) {
  let [user, errWho] = await who();
  if (errWho) {
    return errWho;
  }

  let [res, errPost] = await withErr(fetch(
    `/api/v1/users/${user.id}/twofactor/otp`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      redirect: "follow",
      credentials: "include",
      body: JSON.stringify(body),
    },
  ));
  if (errPost) {
    return errPost;
  }

  if ([400, 404, 500].includes(res.status)) {
    let httpErr = new HTTPError("Failed to add new OTP to account.");
    return httpErr;
  }

  return null;
}

async function twoFactorMethods(userID) {
  let uri = `/api/v1/users/${userID}/twofactor`;
  let [res, errRes] = await withErr(fetch(uri, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  }));
  if (errRes) {
    return [null, errRes];
  }

  if ([400, 401, 404, 500].includes(res.status)) {
    let httpErr = new HTTPError("Failed to remove two factor method.");
    return httpErr;
  }

  let [jsonRes, errJson] = await withErr(res.json());
  if (errJson) {
    return [null, errJson];
  }

  return [jsonRes, null];
}

async function removeTwoFactorMethod(userID, twoFactorID) {
  let uri = `/api/v1/users/${userID}/twofactor/${twoFactorID}`;
  let [res, errDel] = await withErr(fetch(uri, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  }));
  if (errDel) {
    return [null, errDel];
  }
  if ([400, 401, 404, 500].includes(res.status)) {
    let httpErr = new HTTPError("Failed to remove two factor method.");
    return httpErr;
  }

  return null;
}

export { newOTP, optionsOTP, updatePassword, who, twoFactorMethods, removeTwoFactorMethod };
